package multiminer

import (
	"context"
	"testing"
	"time"
)

// mockDriver for testing
type mockDriver struct {
	name         string
	shouldDetect bool
	detectError  error
}

func (d *mockDriver) Name() string { return d.name }
func (d *mockDriver) Detect(ctx context.Context, ep Endpoint) (bool, error) {
	if d.detectError != nil {
		return false, d.detectError
	}
	return d.shouldDetect, nil
}
func (d *mockDriver) Capabilities() Capability {
	return Capability{ReadStats: true, ReadSummary: true}
}
func (d *mockDriver) Open(ctx context.Context, ep Endpoint) (Session, error) {
	return &mockSession{}, nil
}

// mockSession for testing
type mockSession struct{}

func (s *mockSession) Close() error                                                   { return nil }
func (s *mockSession) Model(ctx context.Context) (Model, error)                      { return Model{Vendor: "Mock"}, nil }
func (s *mockSession) Stats(ctx context.Context) (Stats, error)                      { return Stats{}, nil }
func (s *mockSession) Summary(ctx context.Context) (Summary, error)                  { return Summary{}, nil }
func (s *mockSession) Pools(ctx context.Context) ([]Pool, error)                     { return nil, nil }
func (s *mockSession) AddPool(ctx context.Context, url, user, pass string) error     { return nil }
func (s *mockSession) EnablePool(ctx context.Context, poolID int64) error            { return nil }
func (s *mockSession) DisablePool(ctx context.Context, poolID int64) error           { return nil }
func (s *mockSession) RemovePool(ctx context.Context, poolID int64) error            { return nil }
func (s *mockSession) SwitchPool(ctx context.Context, poolID int64) error            { return nil }
func (s *mockSession) Restart(ctx context.Context) error                             { return nil }
func (s *mockSession) Quit(ctx context.Context) error                                { return nil }
func (s *mockSession) Exec(ctx context.Context, command string, parameter string) ([]byte, error) {
	return []byte("{}"), nil
}
func (s *mockSession) GetPowerMode(ctx context.Context) (PowerMode, error) { return PowerMode{}, nil }
func (s *mockSession) SetPowerMode(ctx context.Context, mode PowerMode) error { return nil }
func (s *mockSession) GetFan(ctx context.Context) (FanConfig, error)         { return FanConfig{}, nil }
func (s *mockSession) SetFan(ctx context.Context, fan FanConfig) error       { return nil }

func TestManagerAddDevice(t *testing.T) {
	reg := NewRegistry()
	driver := &mockDriver{name: "test-driver", shouldDetect: true}
	reg.Register(driver)
	
	mgr := NewManager(reg)
	defer mgr.Close()
	
	ctx := context.Background()
	id := MinerID("test-device")
	ep := Endpoint{Address: "192.168.1.100:4028"}
	
	// Test adding device with specific driver
	err := mgr.AddOrDetect(ctx, id, ep, driver)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	devices := mgr.List()
	if len(devices) != 1 {
		t.Errorf("Expected 1 device, got %d", len(devices))
	}
	
	if devices[0].ID != id {
		t.Errorf("Expected device ID %s, got %s", id, devices[0].ID)
	}
}

func TestManagerAutoDetect(t *testing.T) {
	reg := NewRegistry()
	driver := &mockDriver{name: "auto-driver", shouldDetect: true}
	reg.Register(driver)
	
	mgr := NewManager(reg)
	defer mgr.Close()
	
	ctx := context.Background()
	id := MinerID("auto-device")
	ep := Endpoint{Address: "192.168.1.101:4028"}
	
	// Test auto-detection
	err := mgr.AddOrDetect(ctx, id, ep, nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	devices := mgr.List()
	if len(devices) != 1 {
		t.Errorf("Expected 1 device, got %d", len(devices))
	}
	
	if devices[0].DriverName != "auto-driver" {
		t.Errorf("Expected driver name %s, got %s", "auto-driver", devices[0].DriverName)
	}
}

func TestManagerWithSession(t *testing.T) {
	reg := NewRegistry()
	driver := &mockDriver{name: "session-driver", shouldDetect: true}
	reg.Register(driver)
	
	mgr := NewManager(reg)
	defer mgr.Close()
	
	ctx := context.Background()
	id := MinerID("session-device")
	ep := Endpoint{Address: "192.168.1.102:4028"}
	
	// Add device
	err := mgr.AddOrDetect(ctx, id, ep, driver)
	if err != nil {
		t.Errorf("Expected no error adding device, got %v", err)
	}
	
	// Test session usage
	sessionUsed := false
	err = mgr.WithSession(ctx, id, func(sess Session) error {
		sessionUsed = true
		_, err := sess.Model(ctx)
		return err
	})
	
	if err != nil {
		t.Errorf("Expected no error in session, got %v", err)
	}
	
	if !sessionUsed {
		t.Error("Session callback was not called")
	}
}

func TestManagerDeviceNotFound(t *testing.T) {
	reg := NewRegistry()
	mgr := NewManager(reg)
	defer mgr.Close()
	
	ctx := context.Background()
	id := MinerID("nonexistent")
	
	err := mgr.WithSession(ctx, id, func(sess Session) error {
		return nil
	})
	
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestConnectionPool(t *testing.T) {
	pool := NewConnectionPool()
	defer pool.Close()
	
	pool.SetLimits(2, 5, time.Minute)
	
	// Create mock device
	driver := &mockDriver{name: "pool-driver", shouldDetect: true}
	device := &Device{
		ID:         MinerID("pool-test"),
		Driver:     driver,
		Endpoint:   Endpoint{Address: "192.168.1.103:4028"},
		DriverName: "pool-driver",
	}
	
	ctx := context.Background()
	
	// Get session from pool
	sess1, err := pool.GetSession(ctx, device.ID, device)
	if err != nil {
		t.Errorf("Expected no error getting session, got %v", err)
	}
	
	// Return session to pool
	pool.ReturnSession(device.ID, sess1)
	
	// Get session again (should reuse from pool)
	sess2, err := pool.GetSession(ctx, device.ID, device)
	if err != nil {
		t.Errorf("Expected no error getting session from pool, got %v", err)
	}
	
	if sess1 != sess2 {
		t.Error("Expected to reuse session from pool")
	}
	
	// Check pool stats
	stats := pool.Stats()
	if deviceStats, exists := stats[device.ID]; exists {
		if deviceStats.ActiveConnections != 1 {
			t.Errorf("Expected 1 active connection, got %d", deviceStats.ActiveConnections)
		}
	} else {
		t.Error("Expected device stats to exist")
	}
}