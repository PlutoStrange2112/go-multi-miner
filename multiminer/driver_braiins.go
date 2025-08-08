package multiminer

import (
	"context"
	"net"
	"strings"
	"time"

	cg "github.com/x1unix/go-cgminer-api"
)

// Driver stub for Braiins OS (often cgminer-compatible with extensions).
type braiinsDriver struct{}

func NewBraiinsDriver() Driver        { return &braiinsDriver{} }
func (d *braiinsDriver) Name() string { return "braiins" }
func (d *braiinsDriver) Capabilities() Capability {
	return Capability{ReadStats: true, ReadSummary: true, ListPools: true, ManagePools: true, Restart: true, Quit: true, PowerControl: true, FanControl: true,
		SupportedPowerModes: []PowerModeKind{PowerLow, PowerBalanced, PowerHigh}}
}
func (d *braiinsDriver) Detect(ctx context.Context, ep Endpoint) (bool, error) {
	// Braiins OS typically runs cgminer-compatible API with Braiins-specific version info
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

	// Check for Braiins-specific identifiers in version info
	joined := strings.ToLower(v.Type + " " + v.Miner + " " + v.BMMiner + " " + v.CompileTime)
	if strings.Contains(joined, "braiins") ||
		strings.Contains(joined, "braiinsos") ||
		strings.Contains(joined, "bos") {
		return true, nil
	}

	return false, nil
}

func (d *braiinsDriver) Open(ctx context.Context, ep Endpoint) (Session, error) {
	client := &cg.CGMiner{
		Address:   ep.Address,
		Timeout:   3 * time.Second,
		Transport: cg.NewJSONTransport(),
		Dialer:    &net.Dialer{Timeout: 3 * time.Second},
	}
	return &braiinsSession{c: client}, nil
}

// braiinsSession implements Session for Braiins OS devices
type braiinsSession struct {
	c *cg.CGMiner
}

func (s *braiinsSession) Close() error { return nil }

func (s *braiinsSession) Model(ctx context.Context) (Model, error) {
	v, err := s.c.VersionContext(ctx)
	if err != nil {
		return Model{}, err
	}
	return Model{Vendor: "Braiins", Product: v.Miner, Firmware: "BraiinsOS " + v.BMMiner}, nil
}

func (s *braiinsSession) Stats(ctx context.Context) (Stats, error) {
	st, err := s.c.StatsContext(ctx)
	if err != nil {
		return Stats{}, err
	}
	g := st.Generic()
	return Stats{
		Model:      Model{Vendor: "Braiins", Product: g.Miner, Firmware: g.BMMiner},
		Hashrate5s: g.Ghs5s.Float64(),
		HashrateAv: g.GhsAverage,
		TempMax:    float64(g.TempMax),
		UptimeSec:  g.Elapsed,
	}, nil
}

func (s *braiinsSession) Summary(ctx context.Context) (Summary, error) {
	sm, err := s.c.SummaryContext(ctx)
	if err != nil {
		return Summary{}, err
	}
	return Summary{
		Accepted:              sm.Accepted,
		Rejected:              sm.Rejected,
		DeviceHardwarePercent: sm.DeviceHardwarePercent,
		GHS5s:                 sm.GHS5s.Float64(),
		GHSav:                 sm.GHSav,
	}, nil
}

func (s *braiinsSession) Pools(ctx context.Context) ([]Pool, error) {
	pls, err := s.c.PoolsContext(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Pool, 0, len(pls))
	for _, p := range pls {
		out = append(out, Pool{ID: p.Pool, URL: p.URL, User: p.User, Priority: p.Priority, Active: p.StratumActive})
	}
	return out, nil
}

func (s *braiinsSession) AddPool(ctx context.Context, url, user, pass string) error {
	return s.c.AddPoolContext(ctx, url, user, pass)
}

func (s *braiinsSession) EnablePool(ctx context.Context, poolID int64) error {
	return s.c.EnablePoolContext(ctx, &cg.Pool{Pool: poolID})
}

func (s *braiinsSession) DisablePool(ctx context.Context, poolID int64) error {
	return s.c.DisablePoolContext(ctx, &cg.Pool{Pool: poolID})
}

func (s *braiinsSession) RemovePool(ctx context.Context, poolID int64) error {
	return NewDeviceError("remove pool not implemented", "Braiins OS may not support pool removal", nil)
}

func (s *braiinsSession) SwitchPool(ctx context.Context, poolID int64) error {
	return NewDeviceError("switch pool not implemented", "Braiins OS may not support pool switching", nil)
}

func (s *braiinsSession) Restart(ctx context.Context) error {
	return s.c.CallContext(ctx, cg.NewCommandWithoutParameter("restart"), nil)
}

func (s *braiinsSession) Quit(ctx context.Context) error {
	return s.c.CallContext(ctx, cg.NewCommandWithoutParameter("quit"), nil)
}

func (s *braiinsSession) Exec(ctx context.Context, command string, parameter string) ([]byte, error) {
	return s.c.RawCall(ctx, cg.NewCommand(command, parameter))
}

// Power management - Braiins OS supports power tuning
func (s *braiinsSession) GetPowerMode(ctx context.Context) (PowerMode, error) {
	// This is a stub - in reality you'd query Braiins-specific APIs
	return PowerMode{Kind: PowerBalanced}, nil
}

func (s *braiinsSession) SetPowerMode(ctx context.Context, mode PowerMode) error {
	// This is a stub - would use Braiins-specific power tuning APIs
	return NewDeviceError("power mode setting not fully implemented", "would require Braiins-specific API calls", nil)
}

func (s *braiinsSession) GetFan(ctx context.Context) (FanConfig, error) {
	return FanConfig{Mode: FanAuto}, nil
}

func (s *braiinsSession) SetFan(ctx context.Context, fan FanConfig) error {
	return NewDeviceError("fan control not fully implemented", "would require Braiins-specific API calls", nil)
}
