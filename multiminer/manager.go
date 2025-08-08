package multiminer

import (
	"context"
	"sync"
	"time"
)

// Device represents a tracked miner.
type Device struct {
	ID         MinerID
	Endpoint   Endpoint
	Driver     Driver
	DriverName string
}

// Manager tracks devices and provides operations across them.
type Manager struct {
	reg  *Registry
	mu   sync.RWMutex
	dev  map[MinerID]*Device
	opt  ManagerOptions
	pool *ConnectionPool
}

func NewManager(reg *Registry) *Manager { return NewManagerWithOptions(reg, defaultOptions()) }
func NewManagerWithOptions(reg *Registry, opt ManagerOptions) *Manager {
	pool := NewConnectionPool()
	pool.SetLimits(5, 10, 5*time.Minute)

	return &Manager{
		reg:  reg,
		dev:  make(map[MinerID]*Device),
		opt:  opt,
		pool: pool,
	}
}

// AddOrDetect registers a device, auto-detecting driver if needed.
func (m *Manager) AddOrDetect(ctx context.Context, id MinerID, ep Endpoint, d Driver) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if d == nil {
		var err error
		d, err = m.reg.Detect(ctx, ep)
		if err != nil {
			return err
		}
	}
	name := d.Name()
	m.dev[id] = &Device{ID: id, Endpoint: ep, Driver: d, DriverName: name}
	return nil
}

func (m *Manager) List() []Device {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Device, 0, len(m.dev))
	for _, d := range m.dev {
		out = append(out, *d)
	}
	return out
}

// WithSession opens a session for a device and runs fn using connection pool.
func (m *Manager) WithSession(ctx context.Context, id MinerID, fn func(Session) error) error {
	m.mu.RLock()
	d := m.dev[id]
	m.mu.RUnlock()
	if d == nil {
		return ErrNotFound
	}

	sess, err := m.pool.GetSession(ctx, id, d)
	if err != nil {
		return err
	}

	// Ensure session is returned to pool
	defer m.pool.ReturnSession(id, sess)

	return fn(sess)
}

// DeviceInfo is a safe DTO for API responses.
type DeviceInfo struct {
	ID      string `json:"id"`
	Address string `json:"address"`
	Driver  string `json:"driver"`
}

func (m *Manager) DeviceInfos() []DeviceInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]DeviceInfo, 0, len(m.dev))
	for _, d := range m.dev {
		out = append(out, DeviceInfo{ID: string(d.ID), Address: d.Endpoint.Address, Driver: d.DriverName})
	}
	return out
}

// Close gracefully shuts down the manager and connection pool
func (m *Manager) Close() error {
	m.pool.Close()
	return nil
}

// StartCleanup starts background cleanup of expired connections
func (m *Manager) StartCleanup(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				m.pool.CleanUp()
			}
		}
	}()
}

// GetPoolStats returns connection pool statistics
func (m *Manager) GetPoolStats() map[MinerID]PoolStats {
	return m.pool.Stats()
}
