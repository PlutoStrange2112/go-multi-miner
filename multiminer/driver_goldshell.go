package multiminer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Driver stub for Goldshell.
type goldshellDriver struct{}

func NewGoldshellDriver() Driver { return &goldshellDriver{} }
func (d *goldshellDriver) Name() string { return "goldshell" }
func (d *goldshellDriver) Capabilities() Capability {
	return Capability{ReadStats: true, ReadSummary: true, ListPools: true, ManagePools: true, Restart: true, Quit: true, PowerControl: true, FanControl: true}
}
func (d *goldshellDriver) Detect(ctx context.Context, ep Endpoint) (bool, error) {
	// Goldshell miners typically expose HTTP API on port 80 or 8080
	// Try to detect via HTTP API endpoints
	candidates := []string{"/mcb/status", "/api/status", "/status", "/"}
	
	path, found := probeHTTP(ctx, ep.Address, candidates, 1200*time.Millisecond)
	if !found {
		return false, nil
	}
	
	// Try to get more info from the status endpoint
	client := &http.Client{Timeout: 1200 * time.Millisecond}
	url := fmt.Sprintf("http://%s%s", ep.Address, path)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := client.Do(req)
	if err != nil {
		return true, nil // We found HTTP response, assume it's Goldshell
	}
	defer resp.Body.Close()
	
	// Look for Goldshell-specific indicators in response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return true, nil // HTTP response found, assume Goldshell
	}
	
	// Check for Goldshell-specific keys in JSON response
	respStr := strings.ToLower(fmt.Sprintf("%v", result))
	if strings.Contains(respStr, "goldshell") || 
	   strings.Contains(respStr, "kd-box") ||
	   strings.Contains(respStr, "hs-box") {
		return true, nil
	}
	
	return true, nil // If we got a proper JSON response, assume it's Goldshell
}

func (d *goldshellDriver) Open(ctx context.Context, ep Endpoint) (Session, error) {
	return &goldshellSession{address: ep.Address}, nil
}

// goldshellSession implements Session for Goldshell devices
type goldshellSession struct {
	address string
}

func (s *goldshellSession) Close() error { return nil }

func (s *goldshellSession) Model(ctx context.Context) (Model, error) {
	// Try to get device info from HTTP API
	client := &http.Client{Timeout: 3 * time.Second}
	url := fmt.Sprintf("http://%s/api/status", s.address)
	
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := client.Do(req)
	if err != nil {
		return Model{Vendor: "Goldshell", Product: "Unknown", Firmware: "Unknown"}, nil
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Model{Vendor: "Goldshell", Product: "Unknown", Firmware: "Unknown"}, nil
	}
	
	product := "Unknown"
	firmware := "Unknown"
	
	if model, ok := result["model"].(string); ok {
		product = model
	}
	if fw, ok := result["firmware"].(string); ok {
		firmware = fw
	}
	
	return Model{Vendor: "Goldshell", Product: product, Firmware: firmware}, nil
}

func (s *goldshellSession) Stats(ctx context.Context) (Stats, error) {
	model, _ := s.Model(ctx)
	
	client := &http.Client{Timeout: 3 * time.Second}
	url := fmt.Sprintf("http://%s/api/stats", s.address)
	
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := client.Do(req)
	if err != nil {
		return Stats{Model: model}, NewConnectionError("failed to get stats", err)
	}
	defer resp.Body.Close()
	
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Stats{Model: model}, NewDeviceError("failed to parse stats", "invalid JSON response", err)
	}
	
	stats := Stats{Model: model}
	
	if hashrate, ok := result["hashrate"].(float64); ok {
		stats.HashrateAv = hashrate / 1000000000 // Convert to GH/s
		stats.Hashrate5s = stats.HashrateAv     // Use same value for 5s
	}
	
	if temp, ok := result["temperature"].(float64); ok {
		stats.TempMax = temp
	}
	
	if uptime, ok := result["uptime"].(float64); ok {
		stats.UptimeSec = int64(uptime)
	}
	
	return stats, nil
}

func (s *goldshellSession) Summary(ctx context.Context) (Summary, error) {
	client := &http.Client{Timeout: 3 * time.Second}
	url := fmt.Sprintf("http://%s/api/summary", s.address)
	
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := client.Do(req)
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
	
	if hashrate, ok := result["hashrate"].(float64); ok {
		ghash := hashrate / 1000000000 // Convert to GH/s
		summary.GHSav = ghash
		summary.GHS5s = ghash
	}
	
	return summary, nil
}

func (s *goldshellSession) Pools(ctx context.Context) ([]Pool, error) {
	return nil, NewDeviceError("pools not implemented", "Goldshell pool management via HTTP API not implemented", nil)
}

func (s *goldshellSession) AddPool(ctx context.Context, url, user, pass string) error {
	return NewDeviceError("add pool not implemented", "Goldshell pool management not implemented", nil)
}

func (s *goldshellSession) EnablePool(ctx context.Context, poolID int64) error {
	return NewDeviceError("enable pool not implemented", "Goldshell pool management not implemented", nil)
}

func (s *goldshellSession) DisablePool(ctx context.Context, poolID int64) error {
	return NewDeviceError("disable pool not implemented", "Goldshell pool management not implemented", nil)
}

func (s *goldshellSession) RemovePool(ctx context.Context, poolID int64) error {
	return NewDeviceError("remove pool not implemented", "Goldshell pool management not implemented", nil)
}

func (s *goldshellSession) SwitchPool(ctx context.Context, poolID int64) error {
	return NewDeviceError("switch pool not implemented", "Goldshell pool management not implemented", nil)
}

func (s *goldshellSession) Restart(ctx context.Context) error {
	return NewDeviceError("restart not implemented", "Goldshell restart via HTTP API not implemented", nil)
}

func (s *goldshellSession) Quit(ctx context.Context) error {
	return NewDeviceError("quit not implemented", "Goldshell quit not applicable via HTTP API", nil)
}

func (s *goldshellSession) Exec(ctx context.Context, command string, parameter string) ([]byte, error) {
	return nil, NewDeviceError("exec not implemented", "Goldshell raw command execution not supported", nil)
}

func (s *goldshellSession) GetPowerMode(ctx context.Context) (PowerMode, error) {
	return PowerMode{Kind: PowerBalanced}, NewDeviceError("power mode not implemented", "Goldshell power mode reading not implemented", nil)
}

func (s *goldshellSession) SetPowerMode(ctx context.Context, mode PowerMode) error {
	return NewDeviceError("power mode setting not implemented", "Goldshell power mode setting not implemented", nil)
}

func (s *goldshellSession) GetFan(ctx context.Context) (FanConfig, error) {
	return FanConfig{Mode: FanAuto}, NewDeviceError("fan control not implemented", "Goldshell fan reading not implemented", nil)
}

func (s *goldshellSession) SetFan(ctx context.Context, fan FanConfig) error {
	return NewDeviceError("fan control not implemented", "Goldshell fan control not implemented", nil)
}
