package multiminer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Driver stub for iPollo.
type ipolloDriver struct{}

func NewIPolloDriver() Driver        { return &ipolloDriver{} }
func (d *ipolloDriver) Name() string { return "ipollo" }
func (d *ipolloDriver) Capabilities() Capability {
	return Capability{ReadStats: true, ReadSummary: true, ListPools: true, ManagePools: true, Restart: true, Quit: true, PowerControl: true, FanControl: true}
}

func (d *ipolloDriver) Detect(ctx context.Context, ep Endpoint) (bool, error) {
	// iPollo miners typically expose HTTP API on port 80 or 4028
	// Try to detect via HTTP API endpoints
	candidates := []string{"/api/status", "/cgi-bin/status", "/status", "/"}

	path, found := probeHTTP(ctx, ep.Address, candidates, 1200*time.Millisecond)
	if !found {
		return false, nil
	}

	// Try to get more info from the status endpoint to confirm it's iPollo
	client := &http.Client{Timeout: 1200 * time.Millisecond}
	url := fmt.Sprintf("http://%s%s", ep.Address, path)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := client.Do(req)
	if err != nil {
		return true, nil // We found HTTP response, assume it's iPollo
	}
	defer resp.Body.Close()

	// Look for iPollo-specific indicators in response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return true, nil // HTTP response found, assume iPollo
	}

	// Check for iPollo-specific keys in JSON response
	respStr := strings.ToLower(fmt.Sprintf("%v", result))
	if strings.Contains(respStr, "ipollo") ||
		strings.Contains(respStr, "nanominer") ||
		strings.Contains(respStr, "v1mini") {
		return true, nil
	}

	return true, nil // If we got a proper JSON response, assume it's iPollo
}

func (d *ipolloDriver) Open(ctx context.Context, ep Endpoint) (Session, error) {
	return &ipolloSession{address: ep.Address}, nil
}

// ipolloSession implements Session for iPollo devices
type ipolloSession struct {
	address    string
	httpClient *http.Client
}

func (s *ipolloSession) ensureClient() {
	if s.httpClient == nil {
		s.httpClient = &http.Client{Timeout: 3 * time.Second}
	}
}

func (s *ipolloSession) Close() error { return nil }

func (s *ipolloSession) Model(ctx context.Context) (Model, error) {
	s.ensureClient()

	// Try to get device info from HTTP API
	endpoints := []string{"/api/status", "/cgi-bin/status", "/status"}

	for _, endpoint := range endpoints {
		url := fmt.Sprintf("http://%s%s", s.address, endpoint)
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

		model := Model{Vendor: "iPollo", Product: "Unknown", Firmware: "Unknown"}

		// Extract model information
		if minerType, ok := result["miner_type"].(string); ok {
			model.Product = minerType
		} else if hw, ok := result["hardware"].(string); ok {
			model.Product = hw
		} else if model_name, ok := result["model"].(string); ok {
			model.Product = model_name
		}

		if fw, ok := result["firmware"].(string); ok {
			model.Firmware = fw
		} else if version, ok := result["version"].(string); ok {
			model.Firmware = version
		}

		return model, nil
	}

	return Model{Vendor: "iPollo", Product: "Unknown", Firmware: "Unknown"}, nil
}

func (s *ipolloSession) Stats(ctx context.Context) (Stats, error) {
	s.ensureClient()
	model, _ := s.Model(ctx)

	// Try to get stats from HTTP API
	endpoints := []string{"/api/stats", "/cgi-bin/stats", "/stats", "/api/status"}

	for _, endpoint := range endpoints {
		url := fmt.Sprintf("http://%s%s", s.address, endpoint)
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

		// Extract hashrate information
		if hashrate, ok := result["hashrate"].(float64); ok {
			stats.HashrateAv = hashrate / 1000000000 // Convert to GH/s
			stats.Hashrate5s = stats.HashrateAv      // Use same value for 5s
		} else if hashrateStr, ok := result["hashrate"].(string); ok {
			// Parse hashrate string if needed
			var hr float64
			fmt.Sscanf(hashrateStr, "%f", &hr)
			stats.HashrateAv = hr / 1000000000
			stats.Hashrate5s = stats.HashrateAv
		}

		// Extract temperature
		if temp, ok := result["temp_max"].(float64); ok {
			stats.TempMax = temp
		} else if temp, ok := result["temperature"].(float64); ok {
			stats.TempMax = temp
		} else if temps, ok := result["temps"].([]interface{}); ok && len(temps) > 0 {
			// Find max temperature
			maxTemp := 0.0
			for _, t := range temps {
				if temp, ok := t.(float64); ok && temp > maxTemp {
					maxTemp = temp
				}
			}
			stats.TempMax = maxTemp
		}

		// Extract uptime
		if uptime, ok := result["uptime"].(float64); ok {
			stats.UptimeSec = int64(uptime)
		}

		return stats, nil
	}

	return Stats{Model: model}, NewDeviceError("stats not available", "no working iPollo stats endpoint found", nil)
}

func (s *ipolloSession) Summary(ctx context.Context) (Summary, error) {
	s.ensureClient()

	// Use stats data to build summary
	stats, err := s.Stats(ctx)
	if err != nil {
		return Summary{}, err
	}

	summary := Summary{
		GHS5s: stats.Hashrate5s,
		GHSav: stats.HashrateAv,
	}

	// Try to get pool stats if available
	endpoints := []string{"/api/summary", "/cgi-bin/summary", "/summary"}

	for _, endpoint := range endpoints {
		url := fmt.Sprintf("http://%s%s", s.address, endpoint)
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

		// Extract accepted/rejected shares
		if accepted, ok := result["accepted"].(float64); ok {
			summary.Accepted = int64(accepted)
		}

		if rejected, ok := result["rejected"].(float64); ok {
			summary.Rejected = int64(rejected)
		}

		break
	}

	return summary, nil
}

func (s *ipolloSession) Pools(ctx context.Context) ([]Pool, error) {
	s.ensureClient()

	endpoints := []string{"/api/pools", "/cgi-bin/pools", "/pools"}

	for _, endpoint := range endpoints {
		url := fmt.Sprintf("http://%s%s", s.address, endpoint)
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

		var pools []Pool

		if poolsList, ok := result["pools"].([]interface{}); ok {
			for i, p := range poolsList {
				if poolMap, ok := p.(map[string]interface{}); ok {
					pool := Pool{ID: int64(i)}

					if url, ok := poolMap["url"].(string); ok {
						pool.URL = url
					}

					if user, ok := poolMap["user"].(string); ok {
						pool.User = user
					}

					if priority, ok := poolMap["priority"].(float64); ok {
						pool.Priority = int64(priority)
					}

					if active, ok := poolMap["active"].(bool); ok {
						pool.Active = active
					}

					pools = append(pools, pool)
				}
			}
		}

		return pools, nil
	}

	return nil, NewDeviceError("pools not available", "no working iPollo pools endpoint found", nil)
}

func (s *ipolloSession) AddPool(ctx context.Context, url, user, pass string) error {
	return NewDeviceError("add pool not implemented", "iPollo pool management not yet implemented", nil)
}

func (s *ipolloSession) EnablePool(ctx context.Context, poolID int64) error {
	return NewDeviceError("enable pool not implemented", "iPollo pool management not yet implemented", nil)
}

func (s *ipolloSession) DisablePool(ctx context.Context, poolID int64) error {
	return NewDeviceError("disable pool not implemented", "iPollo pool management not yet implemented", nil)
}

func (s *ipolloSession) RemovePool(ctx context.Context, poolID int64) error {
	return NewDeviceError("remove pool not implemented", "iPollo pool management not yet implemented", nil)
}

func (s *ipolloSession) SwitchPool(ctx context.Context, poolID int64) error {
	return NewDeviceError("switch pool not implemented", "iPollo pool management not yet implemented", nil)
}

func (s *ipolloSession) Restart(ctx context.Context) error {
	s.ensureClient()

	// Try iPollo-specific restart endpoints
	endpoints := []string{"/api/restart", "/cgi-bin/restart", "/restart"}

	for _, endpoint := range endpoints {
		url := fmt.Sprintf("http://%s%s", s.address, endpoint)
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

	return NewDeviceError("restart failed", "no working iPollo restart endpoint found", nil)
}

func (s *ipolloSession) Quit(ctx context.Context) error {
	return NewDeviceError("quit not applicable", "iPollo does not support quit command", nil)
}

func (s *ipolloSession) Exec(ctx context.Context, command string, parameter string) ([]byte, error) {
	return nil, NewDeviceError("exec not supported", "iPollo does not support raw command execution", nil)
}

func (s *ipolloSession) GetPowerMode(ctx context.Context) (PowerMode, error) {
	return PowerMode{Kind: PowerBalanced}, NewDeviceError("power mode not implemented", "iPollo power mode reading not yet implemented", nil)
}

func (s *ipolloSession) SetPowerMode(ctx context.Context, mode PowerMode) error {
	return NewDeviceError("power mode setting not implemented", "iPollo power mode control not yet implemented", nil)
}

func (s *ipolloSession) GetFan(ctx context.Context) (FanConfig, error) {
	return FanConfig{Mode: FanAuto}, NewDeviceError("fan control not implemented", "iPollo fan reading not yet implemented", nil)
}

func (s *ipolloSession) SetFan(ctx context.Context, fan FanConfig) error {
	return NewDeviceError("fan control not implemented", "iPollo fan control not yet implemented", nil)
}
