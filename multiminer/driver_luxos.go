package multiminer

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	cg "github.com/x1unix/go-cgminer-api"
)

// Driver stub for LuxOS (Bitmain fork with HTTP APIs)
type luxOSDriver struct{}

func NewLuxOSDriver() Driver { return &luxOSDriver{} }
func (d *luxOSDriver) Name() string { return "luxos" }
func (d *luxOSDriver) Capabilities() Capability {
	return Capability{ReadStats: true, ReadSummary: true, ListPools: true, ManagePools: true, Restart: true, Quit: true, PowerControl: true, FanControl: true}
}
func (d *luxOSDriver) Detect(ctx context.Context, ep Endpoint) (bool, error) {
	// LuxOS supports both HTTP API and cgminer API
	// Try HTTP API first (port 8080 typically)
	httpFound := d.detectHTTP(ctx, ep.Address)
	if httpFound {
		return true, nil
	}
	
	// Try cgminer API detection with LuxOS-specific heuristics
	c := &cg.CGMiner{
		Address:   ep.Address,
		Timeout:   1200 * time.Millisecond,
		Transport: cg.NewJSONTransport(),
		Dialer:    &net.Dialer{Timeout: 1200 * time.Millisecond},
	}
	
	v, err := c.VersionContext(ctx)
	if err != nil {
		return false, nil
	}
	
	// Check for LuxOS-specific identifiers
	joined := strings.ToLower(v.Type + " " + v.Miner + " " + v.BMMiner + " " + v.CompileTime)
	if strings.Contains(joined, "luxos") || 
	   strings.Contains(joined, "luxor") ||
	   (strings.Contains(joined, "bitmain") && strings.Contains(joined, "lux")) {
		return true, nil
	}
	
	return false, nil
}

func (d *luxOSDriver) detectHTTP(ctx context.Context, address string) bool {
	// Try common LuxOS HTTP endpoints
	httpCandidates := []string{"/api/v1/status", "/luxos/api/status", "/api/status"}
	
	// Extract host without port, then try common HTTP ports
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		host = address
	}
	
	httpAddresses := []string{
		host + ":8080",
		host + ":80",
		host + ":4028", // Some LuxOS installations use this
	}
	
	for _, addr := range httpAddresses {
		if _, found := probeHTTP(ctx, addr, httpCandidates, 800*time.Millisecond); found {
			return true
		}
	}
	
	return false
}

func (d *luxOSDriver) Open(ctx context.Context, ep Endpoint) (Session, error) {
	return &luxOSSession{address: ep.Address}, nil
}

// luxOSSession implements Session for LuxOS devices
type luxOSSession struct {
	address    string
	httpClient *http.Client
	cgClient   *cg.CGMiner
}

func (s *luxOSSession) ensureClients() {
	if s.httpClient == nil {
		s.httpClient = &http.Client{Timeout: 3 * time.Second}
	}
	if s.cgClient == nil {
		s.cgClient = &cg.CGMiner{
			Address:   s.address,
			Timeout:   3 * time.Second,
			Transport: cg.NewJSONTransport(),
			Dialer:    &net.Dialer{Timeout: 3 * time.Second},
		}
	}
}

func (s *luxOSSession) Close() error { return nil }

func (s *luxOSSession) Model(ctx context.Context) (Model, error) {
	s.ensureClients()
	
	// Try HTTP API first
	if model, err := s.getModelHTTP(ctx); err == nil {
		return model, nil
	}
	
	// Fallback to cgminer API
	v, err := s.cgClient.VersionContext(ctx)
	if err != nil {
		return Model{}, NewConnectionError("failed to get device model", err)
	}
	
	return Model{Vendor: "LuxOS", Product: v.Miner, Firmware: v.BMMiner}, nil
}

func (s *luxOSSession) getModelHTTP(ctx context.Context) (Model, error) {
	// Try to find HTTP endpoint
	host, _, err := net.SplitHostPort(s.address)
	if err != nil {
		host = s.address
	}
	
	urls := []string{
		fmt.Sprintf("http://%s:8080/api/v1/status", host),
		fmt.Sprintf("http://%s:80/luxos/api/status", host),
		fmt.Sprintf("http://%s:4028/api/status", host),
	}
	
	for _, url := range urls {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		resp, err := s.httpClient.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			continue
		}
		
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			continue
		}
		
		model := Model{Vendor: "LuxOS", Product: "Unknown", Firmware: "LuxOS"}
		
		if miner, ok := result["miner_type"].(string); ok {
			model.Product = miner
		} else if hw, ok := result["hardware"].(string); ok {
			model.Product = hw
		}
		
		if fw, ok := result["firmware"].(string); ok {
			model.Firmware = fw
		} else if version, ok := result["version"].(string); ok {
			model.Firmware = "LuxOS " + version
		}
		
		return model, nil
	}
	
	return Model{}, fmt.Errorf("no HTTP endpoint found")
}

func (s *luxOSSession) Stats(ctx context.Context) (Stats, error) {
	s.ensureClients()
	model, _ := s.Model(ctx)
	
	// Try HTTP API first
	if stats, err := s.getStatsHTTP(ctx, model); err == nil {
		return stats, nil
	}
	
	// Fallback to cgminer API
	st, err := s.cgClient.StatsContext(ctx)
	if err != nil {
		return Stats{Model: model}, NewConnectionError("failed to get stats", err)
	}
	
	g := st.Generic()
	return Stats{
		Model:      model,
		Hashrate5s: g.Ghs5s.Float64(),
		HashrateAv: g.GhsAverage,
		TempMax:    float64(g.TempMax),
		UptimeSec:  g.Elapsed,
	}, nil
}

func (s *luxOSSession) getStatsHTTP(ctx context.Context, model Model) (Stats, error) {
	host, _, err := net.SplitHostPort(s.address)
	if err != nil {
		host = s.address
	}
	
	urls := []string{
		fmt.Sprintf("http://%s:8080/api/v1/stats", host),
		fmt.Sprintf("http://%s:80/luxos/api/stats", host),
	}
	
	for _, url := range urls {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		resp, err := s.httpClient.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			continue
		}
		
		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			continue
		}
		
		stats := Stats{Model: model}
		
		if hashrate, ok := result["hashrate_5s"].(float64); ok {
			stats.Hashrate5s = hashrate / 1000000000 // Convert to GH/s
		} else if hashrate, ok := result["hashrate"].(float64); ok {
			stats.Hashrate5s = hashrate / 1000000000
		}
		
		if hashrateAvg, ok := result["hashrate_avg"].(float64); ok {
			stats.HashrateAv = hashrateAvg / 1000000000
		} else {
			stats.HashrateAv = stats.Hashrate5s
		}
		
		if temp, ok := result["temp_max"].(float64); ok {
			stats.TempMax = temp
		} else if temp, ok := result["temperature"].(float64); ok {
			stats.TempMax = temp
		}
		
		if uptime, ok := result["uptime"].(float64); ok {
			stats.UptimeSec = int64(uptime)
		}
		
		return stats, nil
	}
	
	return Stats{}, fmt.Errorf("no HTTP stats endpoint found")
}

func (s *luxOSSession) Summary(ctx context.Context) (Summary, error) {
	s.ensureClients()
	
	// Use cgminer API for summary as it's more standardized
	sm, err := s.cgClient.SummaryContext(ctx)
	if err != nil {
		return Summary{}, NewConnectionError("failed to get summary", err)
	}
	
	return Summary{
		Accepted:              sm.Accepted,
		Rejected:              sm.Rejected,
		DeviceHardwarePercent: sm.DeviceHardwarePercent,
		GHS5s:                 sm.GHS5s.Float64(),
		GHSav:                 sm.GHSav,
	}, nil
}

func (s *luxOSSession) Pools(ctx context.Context) ([]Pool, error) {
	s.ensureClients()
	
	pls, err := s.cgClient.PoolsContext(ctx)
	if err != nil {
		return nil, NewConnectionError("failed to get pools", err)
	}
	
	out := make([]Pool, 0, len(pls))
	for _, p := range pls {
		out = append(out, Pool{ID: p.Pool, URL: p.URL, User: p.User, Priority: p.Priority, Active: p.StratumActive})
	}
	return out, nil
}

func (s *luxOSSession) AddPool(ctx context.Context, url, user, pass string) error {
	s.ensureClients()
	return s.cgClient.AddPoolContext(ctx, url, user, pass)
}

func (s *luxOSSession) EnablePool(ctx context.Context, poolID int64) error {
	s.ensureClients()
	return s.cgClient.EnablePoolContext(ctx, &cg.Pool{Pool: poolID})
}

func (s *luxOSSession) DisablePool(ctx context.Context, poolID int64) error {
	s.ensureClients()
	return s.cgClient.DisablePoolContext(ctx, &cg.Pool{Pool: poolID})
}

func (s *luxOSSession) RemovePool(ctx context.Context, poolID int64) error {
	s.ensureClients()
	return s.cgClient.CallContext(ctx, cg.NewCommand("removepool", fmt.Sprint(poolID)), nil)
}

func (s *luxOSSession) SwitchPool(ctx context.Context, poolID int64) error {
	s.ensureClients()
	return s.cgClient.CallContext(ctx, cg.NewCommand("switchpool", fmt.Sprint(poolID)), nil)
}

func (s *luxOSSession) Restart(ctx context.Context) error {
	s.ensureClients()
	return s.cgClient.CallContext(ctx, cg.NewCommandWithoutParameter("restart"), nil)
}

func (s *luxOSSession) Quit(ctx context.Context) error {
	s.ensureClients()
	return s.cgClient.CallContext(ctx, cg.NewCommandWithoutParameter("quit"), nil)
}

func (s *luxOSSession) Exec(ctx context.Context, command string, parameter string) ([]byte, error) {
	s.ensureClients()
	return s.cgClient.RawCall(ctx, cg.NewCommand(command, parameter))
}

// Power management - LuxOS supports advanced power tuning
func (s *luxOSSession) GetPowerMode(ctx context.Context) (PowerMode, error) {
	s.ensureClients()
	
	// Try HTTP API for power mode
	host, _, err := net.SplitHostPort(s.address)
	if err != nil {
		host = s.address
	}
	
	url := fmt.Sprintf("http://%s:8080/api/v1/power", host)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return PowerMode{Kind: PowerBalanced}, NewDeviceError("power mode not available", "HTTP API not accessible", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return PowerMode{Kind: PowerBalanced}, nil // Default fallback
	}
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return PowerMode{Kind: PowerBalanced}, nil
	}
	
	mode := PowerMode{Kind: PowerBalanced}
	if modeStr, ok := result["mode"].(string); ok {
		switch strings.ToLower(modeStr) {
		case "low", "eco":
			mode.Kind = PowerLow
		case "high", "turbo":
			mode.Kind = PowerHigh
		case "custom":
			mode.Kind = PowerCustom
		default:
			mode.Kind = PowerBalanced
		}
	}
	
	if watts, ok := result["watts"].(float64); ok {
		mode.Watts = int(watts)
	}
	
	return mode, nil
}

func (s *luxOSSession) SetPowerMode(ctx context.Context, mode PowerMode) error {
	s.ensureClients()
	
	host, _, err := net.SplitHostPort(s.address)
	if err != nil {
		host = s.address
	}
	
	// Try HTTP API
	url := fmt.Sprintf("http://%s:8080/api/v1/power", host)
	
	payload := map[string]interface{}{
		"mode": string(mode.Kind),
	}
	
	if mode.Watts > 0 {
		payload["watts"] = mode.Watts
	}
	
	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(jsonData)))
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return NewDeviceError("power mode setting failed", "HTTP API not accessible", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		return NewDeviceError("power mode setting failed", fmt.Sprintf("HTTP %d", resp.StatusCode), nil)
	}
	
	return nil
}

func (s *luxOSSession) GetFan(ctx context.Context) (FanConfig, error) {
	s.ensureClients()
	
	// Default to auto mode
	return FanConfig{Mode: FanAuto}, nil
}

func (s *luxOSSession) SetFan(ctx context.Context, fan FanConfig) error {
	s.ensureClients()
	
	host, _, err := net.SplitHostPort(s.address)
	if err != nil {
		host = s.address
	}
	
	url := fmt.Sprintf("http://%s:8080/api/v1/fans", host)
	
	payload := map[string]interface{}{
		"mode": string(fan.Mode),
	}
	
	if fan.Mode == FanManual {
		payload["speed"] = fan.SpeedPct
	}
	
	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(jsonData)))
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return NewDeviceError("fan control failed", "HTTP API not accessible", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 400 {
		return NewDeviceError("fan control failed", fmt.Sprintf("HTTP %d", resp.StatusCode), nil)
	}
	
	return nil
}
