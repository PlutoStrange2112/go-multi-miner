package multiminer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/x1unix/go-cgminer-api/multiminer/models"
)
// Whatsminer driver using MicroBT HTTP API (skeleton)
type whatsminerDriver struct{}

func NewWhatsminerDriver() Driver { return &whatsminerDriver{} }
func (d *whatsminerDriver) Name() string { return "whatsminer" }
func (d *whatsminerDriver) Capabilities() Capability {
	return Capability{ReadStats: true, ReadSummary: true, ListPools: true, ManagePools: true, Restart: true, Quit: true, PowerControl: true, FanControl: true}
}

func (d *whatsminerDriver) Detect(ctx context.Context, ep Endpoint) (bool, error) {
	// Preference: fastest/lightest first. Try HTTP probe, then fall back (e.g., TCP if available).
	if _, ok := probeHTTP(ctx, ep.Address, []string{"/cgi-bin/minerStatus.cgi", "/api/status", "/"}, 1200*time.Millisecond); ok {
		return true, nil
	}
	// TODO: Try TCP probe to Whatsminer mm API (secondary) if needed.
	return false, nil
}

func (d *whatsminerDriver) Open(ctx context.Context, ep Endpoint) (Session, error) {
	// Decide protocol at open time: prefer HTTP if available, else fallback.
	if path, ok := probeHTTP(ctx, ep.Address, []string{"/api/status", "/cgi-bin/minerStatus.cgi"}, 1200*time.Millisecond); ok {
		return &whatsminerSession{addr: ep.Address, basePath: path, useHTTP: true}, nil
	}
	// Fallback: TCP or alternative
	return &whatsminerSession{addr: ep.Address, useHTTP: false}, nil
}

type whatsminerSession struct{ 
	addr       string
	basePath   string 
	useHTTP    bool
	httpClient *http.Client
}

func (s *whatsminerSession) ensureClient() {
	if s.httpClient == nil {
		s.httpClient = &http.Client{Timeout: 3 * time.Second}
	}
}

func (s *whatsminerSession) Close() error { return nil }

func (s *whatsminerSession) Model(ctx context.Context) (Model, error) {
	if !s.useHTTP {
		return Model{Vendor: "MicroBT", Product: "Whatsminer", Firmware: "Unknown"}, nil
	}
	
	s.ensureClient()
	
	// Try to get device info from status endpoint
	url := fmt.Sprintf("http://%s%s", s.addr, s.basePath)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return Model{Vendor: "MicroBT", Product: "Whatsminer", Firmware: "Unknown"}, nil
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Model{Vendor: "MicroBT", Product: "Whatsminer", Firmware: "Unknown"}, nil
	}
	
	model := Model{Vendor: "MicroBT", Product: "Whatsminer", Firmware: "Unknown"}
	
	// Try to extract model information
	if minerType, ok := result["miner_type"].(string); ok {
		if m, found := models.MatchWhatsminer(minerType); found {
			model.Product = m.Name
		} else {
			model.Product = minerType
		}
	} else if hw, ok := result["hardware"].(string); ok {
		model.Product = hw
	}
	
	if fw, ok := result["firmware"].(string); ok {
		model.Firmware = fw
	} else if version, ok := result["version"].(string); ok {
		model.Firmware = version
	}
	
	return model, nil
}

func (s *whatsminerSession) Stats(ctx context.Context) (Stats, error) {
	if !s.useHTTP {
		return Stats{}, NewDeviceError("stats not available", "TCP mode not implemented", nil)
	}
	
	s.ensureClient()
	model, _ := s.Model(ctx)
	
	url := fmt.Sprintf("http://%s%s", s.addr, s.basePath)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return Stats{Model: model}, NewConnectionError("failed to get stats", err)
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Stats{Model: model}, NewDeviceError("failed to parse stats", "invalid JSON response", err)
	}
	
	stats := Stats{Model: model}
	
	// Extract hashrate information
	if hashrate, ok := result["hashrate_instant"].(float64); ok {
		stats.Hashrate5s = hashrate / 1000000000 // Convert to GH/s
	} else if hashrate, ok := result["hashrate"].(string); ok {
		// Sometimes hashrate comes as string like "95.12 TH/s"
		if parsed := parseHashrateString(hashrate); parsed > 0 {
			stats.Hashrate5s = parsed
		}
	}
	
	if hashrateAvg, ok := result["hashrate_avg"].(float64); ok {
		stats.HashrateAv = hashrateAvg / 1000000000
	} else {
		stats.HashrateAv = stats.Hashrate5s
	}
	
	// Extract temperature
	if temp, ok := result["temp_max"].(float64); ok {
		stats.TempMax = temp
	} else if tempMap, ok := result["temperature"].(map[string]interface{}); ok {
		if maxTemp, ok := tempMap["max"].(float64); ok {
			stats.TempMax = maxTemp
		}
	}
	
	// Extract uptime
	if uptime, ok := result["uptime"].(float64); ok {
		stats.UptimeSec = int64(uptime)
	}
	
	return stats, nil
}

func (s *whatsminerSession) Summary(ctx context.Context) (Summary, error) {
	if !s.useHTTP {
		return Summary{}, NewDeviceError("summary not available", "TCP mode not implemented", nil)
	}
	
	s.ensureClient()
	
	url := fmt.Sprintf("http://%s%s", s.addr, s.basePath)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return Summary{}, NewConnectionError("failed to get summary", err)
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Summary{}, NewDeviceError("failed to parse summary", "invalid JSON response", err)
	}
	
	summary := Summary{}
	
	if accepted, ok := result["accepted"].(float64); ok {
		summary.Accepted = int64(accepted)
	}
	
	if rejected, ok := result["rejected"].(float64); ok {
		summary.Rejected = int64(rejected)
	}
	
	if hashrate, ok := result["hashrate_instant"].(float64); ok {
		ghash := hashrate / 1000000000 // Convert to GH/s
		summary.GHS5s = ghash
		summary.GHSav = ghash
	}
	
	if hashrateAvg, ok := result["hashrate_avg"].(float64); ok {
		summary.GHSav = hashrateAvg / 1000000000
	}
	
	return summary, nil
}

func (s *whatsminerSession) Pools(ctx context.Context) ([]Pool, error) {
	return nil, NewDeviceError("pool management not implemented", "Whatsminer pool management via HTTP not yet implemented", nil)
}

func (s *whatsminerSession) AddPool(ctx context.Context, url, user, pass string) error {
	return NewDeviceError("add pool not implemented", "Whatsminer pool management not yet implemented", nil)
}

func (s *whatsminerSession) EnablePool(ctx context.Context, poolID int64) error {
	return NewDeviceError("enable pool not implemented", "Whatsminer pool management not yet implemented", nil)
}

func (s *whatsminerSession) DisablePool(ctx context.Context, poolID int64) error {
	return NewDeviceError("disable pool not implemented", "Whatsminer pool management not yet implemented", nil)
}

func (s *whatsminerSession) RemovePool(ctx context.Context, poolID int64) error {
	return NewDeviceError("remove pool not implemented", "Whatsminer pool management not yet implemented", nil)
}

func (s *whatsminerSession) SwitchPool(ctx context.Context, poolID int64) error {
	return NewDeviceError("switch pool not implemented", "Whatsminer pool management not yet implemented", nil)
}

func (s *whatsminerSession) Restart(ctx context.Context) error {
	if !s.useHTTP {
		return NewDeviceError("restart not available", "TCP mode not implemented", nil)
	}
	
	s.ensureClient()
	
	// Try common restart endpoints
	endpoints := []string{"/cgi-bin/restart.cgi", "/api/restart"}
	
	for _, endpoint := range endpoints {
		url := fmt.Sprintf("http://%s%s", s.addr, endpoint)
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
		resp, err := s.httpClient.Do(req)
		if err != nil {
			continue
		}
		resp.Body.Close()
		
		if resp.StatusCode < 400 {
			return nil // Success
		}
	}
	
	return NewDeviceError("restart failed", "no working restart endpoint found", nil)
}

func (s *whatsminerSession) Quit(ctx context.Context) error {
	return NewDeviceError("quit not applicable", "Whatsminer does not support quit command", nil)
}

func (s *whatsminerSession) Exec(ctx context.Context, command string, parameter string) ([]byte, error) {
	return nil, NewDeviceError("exec not supported", "Whatsminer does not support raw command execution", nil)
}

func (s *whatsminerSession) GetPowerMode(ctx context.Context) (PowerMode, error) {
	if !s.useHTTP {
		return PowerMode{Kind: PowerBalanced}, NewDeviceError("power mode not available", "TCP mode not implemented", nil)
	}
	
	s.ensureClient()
	
	// Try to get power mode from status
	url := fmt.Sprintf("http://%s%s", s.addr, s.basePath)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return PowerMode{Kind: PowerBalanced}, nil // Default fallback
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return PowerMode{Kind: PowerBalanced}, nil
	}
	
	mode := PowerMode{Kind: PowerBalanced}
	
	// Look for power mode indicators
	if powerMode, ok := result["power_mode"].(string); ok {
		switch strings.ToLower(powerMode) {
		case "low", "eco":
			mode.Kind = PowerLow
		case "high", "turbo", "performance":
			mode.Kind = PowerHigh
		case "custom":
			mode.Kind = PowerCustom
		}
	}
	
	if power, ok := result["power_consumption"].(float64); ok {
		mode.Watts = int(power)
	}
	
	return mode, nil
}

func (s *whatsminerSession) SetPowerMode(ctx context.Context, mode PowerMode) error {
	return NewDeviceError("power mode setting not implemented", "Whatsminer power mode control not yet implemented", nil)
}

func (s *whatsminerSession) GetFan(ctx context.Context) (FanConfig, error) {
	return FanConfig{Mode: FanAuto}, NewDeviceError("fan control not implemented", "Whatsminer fan reading not yet implemented", nil)
}

func (s *whatsminerSession) SetFan(ctx context.Context, fan FanConfig) error {
	return NewDeviceError("fan control not implemented", "Whatsminer fan control not yet implemented", nil)
}

// parseHashrateString parses hashrate strings like "95.12 TH/s" to GH/s
func parseHashrateString(hashrate string) float64 {
	// Use regex to extract number and unit
	re := regexp.MustCompile(`([\d\.]+)\s*([KMGT]?)H/s`)
	matches := re.FindStringSubmatch(hashrate)
	if len(matches) < 3 {
		return 0
	}
	
	var value float64
	fmt.Sscanf(matches[1], "%f", &value)
	
	// Convert to GH/s based on unit
	switch matches[2] {
	case "T":
		return value * 1000 // TH/s to GH/s
	case "K":
		return value / 1000 // KH/s to GH/s
	case "M":
		return value // MH/s to GH/s (approximately)
	case "", "G":
		return value // GH/s
	default:
		return value
	}
}
