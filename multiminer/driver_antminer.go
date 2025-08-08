package multiminer

import (
	"context"
	"net"
	"strings"
	"time"

	cg "github.com/x1unix/go-cgminer-api"
	"github.com/x1unix/go-cgminer-api/multiminer/models"
)

// Antminer (Bitmain stock BMminer) driver that composes the cgminer driver with vendor detection.
type antminerDriver struct{ base Driver }

func NewAntminerDriver() Driver { return &antminerDriver{base: NewCGMinerDriver()} }

func (d *antminerDriver) Name() string { return "antminer" }

func (d *antminerDriver) Capabilities() Capability { return d.base.Capabilities() }

func (d *antminerDriver) Detect(ctx context.Context, ep Endpoint) (bool, error) {
	// Query version via cgminer to confirm vendor/model.
	c := &cg.CGMiner{
		Address:  ep.Address,
		Timeout:  1200 * time.Millisecond,
		Transport: cg.NewJSONTransport(),
		Dialer:   &net.Dialer{Timeout: 1200 * time.Millisecond},
	}
	v, err := c.VersionContext(ctx)
	if err != nil {
		return false, nil
	}
	// Heuristics: Antminer devices typically report Type or Miner including "Antminer" / "BMMiner" / Bitmain identifiers.
	joined := strings.ToLower(v.Type + " " + v.Miner + " " + v.BMMiner)
	if strings.Contains(joined, "antminer") || strings.Contains(joined, "bmminer") || strings.Contains(joined, "bitmain") {
		return true, nil
	}
	return false, nil
}

func (d *antminerDriver) Open(ctx context.Context, ep Endpoint) (Session, error) {
	// Wrap the base session to override Model reporting
	bs, err := d.base.Open(ctx, ep)
	if err != nil { return nil, err }
	return &antminerSession{Session: bs}, nil
}

type antminerSession struct{ Session }

func (s *antminerSession) Model(ctx context.Context) (Model, error) {
	// Try to parse model using version description
	// Call through to underlying cgminer Version via Exec to fetch info
	if cgSess, ok := s.Session.(*cgSession); ok {
		v, err := cgSess.c.VersionContext(ctx)
		if err == nil {
			desc := v.Type + " " + v.Miner + " " + v.BMMiner
			if m, ok := models.MatchAntminer(desc); ok {
				return Model{Vendor: "Bitmain", Product: m.Name, Firmware: v.BMMiner}, nil
			}
			return Model{Vendor: "Bitmain", Product: v.Miner, Firmware: v.BMMiner}, nil
		}
	}
	return s.Session.Model(ctx)
}
