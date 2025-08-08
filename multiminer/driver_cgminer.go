package multiminer

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	cg "github.com/x1unix/go-cgminer-api"
)

// cgminerDriver adapts cgminer/BMminer JSON API devices.
type cgminerDriver struct{}

func NewCGMinerDriver() Driver { return &cgminerDriver{} }

func (d *cgminerDriver) Name() string { return "cgminer" }

func (d *cgminerDriver) Capabilities() Capability {
	return Capability{
		ReadStats:   true,
		ReadSummary: true,
		ListPools:   true,
		ManagePools: true,
		Restart:     true,
		Quit:        true,
		Commands: []string{
			"version", "summary", "devs", "pools", "stats", "addpool", "enablepool", "disablepool", "removepool", "switchpool", "restart", "quit",
		},
		// cgminer variants may have vendor-specific power/fan, default to false here
		FanControl:   false,
		PowerControl: false,
	}
}

// Detect tries a lightweight Version call with a short timeout.
func (d *cgminerDriver) Detect(ctx context.Context, ep Endpoint) (bool, error) {
	// Attempt to open and call Version; keep short timeout to avoid blocking.
	c := &cg.CGMiner{
		Address:   ep.Address,
		Timeout:   1200 * time.Millisecond,
		Transport: cg.NewJSONTransport(),
		Dialer:    &net.Dialer{Timeout: 1200 * time.Millisecond},
	}
	// RawCall to avoid strict parsing; just see if we get bytes back.
	_, err := c.RawCall(ctx, cg.NewCommandWithoutParameter("version"))
	if err != nil {
		var connErr cg.ConnectError
		if errors.As(err, &connErr) {
			return false, nil
		}
		// Other errors (like parse) still indicate speaking the protocol.
		return true, nil
	}
	return true, nil
}

// cgSession implements Session backed by cgminer client.
type cgSession struct{ c *cg.CGMiner }

func (d *cgminerDriver) Open(ctx context.Context, ep Endpoint) (Session, error) {
	client := &cg.CGMiner{
		Address:   ep.Address,
		Timeout:   3 * time.Second,
		Transport: cg.NewJSONTransport(),
		Dialer:    &net.Dialer{Timeout: 3 * time.Second},
	}
	return &cgSession{c: client}, nil
}

func (s *cgSession) Close() error { return nil }

func (s *cgSession) Model(ctx context.Context) (Model, error) {
	v, err := s.c.VersionContext(ctx)
	if err != nil {
		return Model{}, err
	}
	return Model{Vendor: "Bitmain/CGMiner", Product: v.Miner, Firmware: v.BMMiner}, nil
}

func (s *cgSession) Stats(ctx context.Context) (Stats, error) {
	st, err := s.c.StatsContext(ctx)
	if err != nil {
		return Stats{}, err
	}
	g := st.Generic()
	return Stats{
		Model:      Model{Vendor: g.Type, Product: g.Miner, Firmware: g.BMMiner},
		Hashrate5s: g.Ghs5s.Float64(),
		HashrateAv: g.GhsAverage,
		TempMax:    float64(g.TempMax),
		UptimeSec:  g.Elapsed,
	}, nil
}

func (s *cgSession) Summary(ctx context.Context) (Summary, error) {
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

func (s *cgSession) Pools(ctx context.Context) ([]Pool, error) {
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

func (s *cgSession) AddPool(ctx context.Context, url, user, pass string) error {
	return s.c.AddPoolContext(ctx, url, user, pass)
}
func (s *cgSession) EnablePool(ctx context.Context, poolID int64) error {
	return s.c.EnablePoolContext(ctx, &cg.Pool{Pool: poolID})
}
func (s *cgSession) DisablePool(ctx context.Context, poolID int64) error {
	return s.c.DisablePoolContext(ctx, &cg.Pool{Pool: poolID})
}
func (s *cgSession) RemovePool(ctx context.Context, poolID int64) error {
	return s.c.CallContext(ctx, cg.NewCommand("removepool", fmt.Sprint(poolID)), nil)
}
func (s *cgSession) SwitchPool(ctx context.Context, poolID int64) error {
	return s.c.CallContext(ctx, cg.NewCommand("switchpool", fmt.Sprint(poolID)), nil)
}
func (s *cgSession) Restart(ctx context.Context) error {
	return s.c.CallContext(ctx, cg.NewCommandWithoutParameter("restart"), nil)
}
func (s *cgSession) Quit(ctx context.Context) error {
	return s.c.CallContext(ctx, cg.NewCommandWithoutParameter("quit"), nil)
}

func (s *cgSession) Exec(ctx context.Context, command string, parameter string) ([]byte, error) {
	return s.c.RawCall(ctx, cg.NewCommand(command, parameter))
}

func (s *cgSession) GetPowerMode(ctx context.Context) (PowerMode, error) {
	return PowerMode{}, NewDeviceError("power control not supported", "cgminer driver does not support power management", nil)
}
func (s *cgSession) SetPowerMode(ctx context.Context, mode PowerMode) error {
	return NewDeviceError("power control not supported", "cgminer driver does not support power management", nil)
}
func (s *cgSession) GetFan(ctx context.Context) (FanConfig, error) {
	return FanConfig{}, NewDeviceError("fan control not supported", "cgminer driver does not support fan management", nil)
}
func (s *cgSession) SetFan(ctx context.Context, fan FanConfig) error {
	return NewDeviceError("fan control not supported", "cgminer driver does not support fan management", nil)
}
