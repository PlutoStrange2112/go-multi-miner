package multiminer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Driver stub for HiveOS (local agent)
type hiveOSDriver struct{}

func NewHiveOSDriver() Driver        { return &hiveOSDriver{} }
func (d *hiveOSDriver) Name() string { return "hiveos" }
func (d *hiveOSDriver) Capabilities() Capability {
	return Capability{ReadStats: true, ReadSummary: true, ListPools: true, ManagePools: true, Restart: true, Quit: true, PowerControl: true, FanControl: true}
}
func (d *hiveOSDriver) Detect(ctx context.Context, ep Endpoint) (bool, error) {
	// HiveOS typically exposes local APIs on common ports
	// Try to detect HiveOS-specific endpoints
	candidates := []string{
		"/hive/v1/stats",
		"/api/v1/stats",
		"/hiveos/stats",
		"/agent/stats",
	}

	if _, found := probeHTTP(ctx, ep.Address, candidates, 1200*time.Millisecond); found {
		return true, nil
	}

	// Also try to detect by looking for HiveOS-specific response patterns
	client := &http.Client{Timeout: 1200 * time.Millisecond}

	// Try common status endpoints
	statusUrls := []string{
		fmt.Sprintf("http://%s/", ep.Address),
		fmt.Sprintf("http://%s/api/status", ep.Address),
	}

	for _, url := range statusUrls {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		resp, err := client.Do(req)
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

		// Look for HiveOS-specific identifiers
		respStr := strings.ToLower(fmt.Sprintf("%v", result))
		if strings.Contains(respStr, "hiveos") ||
			strings.Contains(respStr, "hive") ||
			strings.Contains(respStr, "agent") {
			return true, nil
		}
	}

	return false, nil
}

func (d *hiveOSDriver) Open(ctx context.Context, ep Endpoint) (Session, error) {
	return &hiveOSSession{address: ep.Address}, nil
}

// hiveOSSession implements Session for HiveOS devices
type hiveOSSession struct {
	address    string
	httpClient *http.Client
}

func (s *hiveOSSession) ensureClient() {
	if s.httpClient == nil {
		s.httpClient = &http.Client{Timeout: 3 * time.Second}
	}
}

func (s *hiveOSSession) Close() error { return nil }

func (s *hiveOSSession) Model(ctx context.Context) (Model, error) {
	s.ensureClient()

	// Try to get device info from various HiveOS endpoints
	endpoints := []string{
		"/hive/v1/info",
		"/api/v1/info",
		"/hiveos/info",
		"/agent/info",
		"/api/status",
	}

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

		model := Model{Vendor: "HiveOS", Product: "Unknown", Firmware: "HiveOS"}

		// Extract model information
		if hw, ok := result["hardware"].(string); ok {
			model.Product = hw
		} else if minerType, ok := result["miner_type"].(string); ok {
			model.Product = minerType
		} else if board, ok := result["board"].(string); ok {
			model.Product = board
		}

		if fw, ok := result["firmware"].(string); ok {
			model.Firmware = fw
		} else if version, ok := result["version"].(string); ok {
			model.Firmware = "HiveOS " + version
		} else if hiveVersion, ok := result["hive_version"].(string); ok {
			model.Firmware = "HiveOS " + hiveVersion
		}

		return model, nil
	}

	return Model{Vendor: "HiveOS", Product: "Unknown", Firmware: "HiveOS"}, nil
}

func (s *hiveOSSession) Stats(ctx context.Context) (Stats, error) {
	s.ensureClient()
	model, _ := s.Model(ctx)

	// Try to get stats from HiveOS endpoints
	endpoints := []string{
		"/hive/v1/stats",
		"/api/v1/stats",
		"/hiveos/stats",
		"/agent/stats",
		"/api/stats",
	}

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

		// Extract hashrate (HiveOS may report in various formats)
		if miners, ok := result["miners"].([]interface{}); ok && len(miners) > 0 {
			// Aggregate hashrate from all miners
			totalHashrate := 0.0
			for _, miner := range miners {
				if minerMap, ok := miner.(map[string]interface{}); ok {
					if hr, ok := minerMap["hashrate"].(float64); ok {
						totalHashrate += hr
					} else if hrStr, ok := minerMap["hashrate"].(string); ok {
						// Parse hashrate string if needed
						var hr float64
						fmt.Sscanf(hrStr, "%f", &hr)
						totalHashrate += hr
					}
				}
			}
			stats.Hashrate5s = totalHashrate / 1000000000 // Convert to GH/s
			stats.HashrateAv = stats.Hashrate5s
		} else if hashrate, ok := result["hashrate"].(float64); ok {
			stats.Hashrate5s = hashrate / 1000000000 // Convert to GH/s
			stats.HashrateAv = stats.Hashrate5s
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

	return Stats{Model: model}, NewDeviceError("stats not available", "no working HiveOS stats endpoint found", nil)
}

func (s *hiveOSSession) Summary(ctx context.Context) (Summary, error) {
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
	endpoints := []string{
		"/hive/v1/pools",
		"/api/v1/pools",
		"/hiveos/pools",
		"/api/pools",
	}

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

func (s *hiveOSSession) Pools(ctx context.Context) ([]Pool, error) {
	s.ensureClient()

	endpoints := []string{
		"/hive/v1/pools",
		"/api/v1/pools",
		"/hiveos/pools",
		"/api/pools",
	}

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

	return nil, NewDeviceError("pools not available", "no working HiveOS pools endpoint found", nil)
}

func (s *hiveOSSession) AddPool(ctx context.Context, url, user, pass string) error {
	return NewDeviceError("add pool not implemented", "HiveOS pool management not yet implemented", nil)
}

func (s *hiveOSSession) EnablePool(ctx context.Context, poolID int64) error {
	return NewDeviceError("enable pool not implemented", "HiveOS pool management not yet implemented", nil)
}

func (s *hiveOSSession) DisablePool(ctx context.Context, poolID int64) error {
	return NewDeviceError("disable pool not implemented", "HiveOS pool management not yet implemented", nil)
}

func (s *hiveOSSession) RemovePool(ctx context.Context, poolID int64) error {
	return NewDeviceError("remove pool not implemented", "HiveOS pool management not yet implemented", nil)
}

func (s *hiveOSSession) SwitchPool(ctx context.Context, poolID int64) error {
	return NewDeviceError("switch pool not implemented", "HiveOS pool management not yet implemented", nil)
}

func (s *hiveOSSession) Restart(ctx context.Context) error {
	s.ensureClient()

	// Try HiveOS-specific restart endpoints
	endpoints := []string{
		"/hive/v1/restart",
		"/api/v1/restart",
		"/hiveos/restart",
		"/agent/restart",
		"/api/restart",
	}

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

	return NewDeviceError("restart failed", "no working HiveOS restart endpoint found", nil)
}

func (s *hiveOSSession) Quit(ctx context.Context) error {
	return NewDeviceError("quit not applicable", "HiveOS does not support quit command", nil)
}

func (s *hiveOSSession) Exec(ctx context.Context, command string, parameter string) ([]byte, error) {
	return nil, NewDeviceError("exec not supported", "HiveOS does not support raw command execution", nil)
}

func (s *hiveOSSession) GetPowerMode(ctx context.Context) (PowerMode, error) {
	return PowerMode{Kind: PowerBalanced}, NewDeviceError("power mode not implemented", "HiveOS power mode reading not yet implemented", nil)
}

func (s *hiveOSSession) SetPowerMode(ctx context.Context, mode PowerMode) error {
	return NewDeviceError("power mode setting not implemented", "HiveOS power mode control not yet implemented", nil)
}

func (s *hiveOSSession) GetFan(ctx context.Context) (FanConfig, error) {
	return FanConfig{Mode: FanAuto}, NewDeviceError("fan control not implemented", "HiveOS fan reading not yet implemented", nil)
}

func (s *hiveOSSession) SetFan(ctx context.Context, fan FanConfig) error {
	return NewDeviceError("fan control not implemented", "HiveOS fan control not yet implemented", nil)
}
