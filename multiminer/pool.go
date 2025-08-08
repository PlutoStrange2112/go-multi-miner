package multiminer

import (
	"context"
	"sync"
	"time"
)

// ConnectionPool manages session connections to devices
type ConnectionPool struct {
	mu       sync.RWMutex
	pools    map[MinerID]*DevicePool
	maxIdle  int
	maxOpen  int
	idleTime time.Duration
}

// DevicePool holds connections for a single device
type DevicePool struct {
	mu        sync.Mutex
	device    *Device
	idle      []Session
	active    map[Session]bool
	createdAt map[Session]time.Time
	maxIdle   int
	maxOpen   int
	idleTime  time.Duration
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool() *ConnectionPool {
	return &ConnectionPool{
		pools:    make(map[MinerID]*DevicePool),
		maxIdle:  5,
		maxOpen:  10,
		idleTime: 5 * time.Minute,
	}
}

// SetLimits configures pool limits
func (p *ConnectionPool) SetLimits(maxIdle, maxOpen int, idleTime time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.maxIdle = maxIdle
	p.maxOpen = maxOpen
	p.idleTime = idleTime
}

// GetSession retrieves a session from the pool or creates a new one
func (p *ConnectionPool) GetSession(ctx context.Context, id MinerID, device *Device) (Session, error) {
	p.mu.Lock()
	pool, exists := p.pools[id]
	if !exists {
		pool = &DevicePool{
			device:    device,
			active:    make(map[Session]bool),
			createdAt: make(map[Session]time.Time),
			maxIdle:   p.maxIdle,
			maxOpen:   p.maxOpen,
			idleTime:  p.idleTime,
		}
		p.pools[id] = pool
	}
	p.mu.Unlock()

	return pool.getSession(ctx)
}

// ReturnSession returns a session to the pool
func (p *ConnectionPool) ReturnSession(id MinerID, sess Session) {
	p.mu.RLock()
	pool, exists := p.pools[id]
	p.mu.RUnlock()

	if exists {
		pool.returnSession(sess)
	} else {
		// Pool doesn't exist anymore, close session
		sess.Close()
	}
}

// CleanUp removes expired idle connections
func (p *ConnectionPool) CleanUp() {
	p.mu.RLock()
	pools := make([]*DevicePool, 0, len(p.pools))
	for _, pool := range p.pools {
		pools = append(pools, pool)
	}
	p.mu.RUnlock()

	for _, pool := range pools {
		pool.cleanExpired()
	}
}

// Close closes all connections and clears the pools
func (p *ConnectionPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, pool := range p.pools {
		pool.closeAll()
	}
	p.pools = make(map[MinerID]*DevicePool)
}

// getSession gets a session from the device pool
func (dp *DevicePool) getSession(ctx context.Context) (Session, error) {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	// Try to get an idle session
	if len(dp.idle) > 0 {
		sess := dp.idle[len(dp.idle)-1]
		dp.idle = dp.idle[:len(dp.idle)-1]
		dp.active[sess] = true
		return sess, nil
	}

	// Check if we can create a new session
	if len(dp.active) >= dp.maxOpen {
		return nil, NewDeviceError("connection pool exhausted", "too many active connections", nil)
	}

	// Create new session
	sess, err := dp.device.Driver.Open(ctx, dp.device.Endpoint)
	if err != nil {
		return nil, err
	}

	dp.active[sess] = true
	dp.createdAt[sess] = time.Now()
	return sess, nil
}

// returnSession returns a session to the device pool
func (dp *DevicePool) returnSession(sess Session) {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	// Remove from active
	delete(dp.active, sess)

	// Add to idle pool if there's space
	if len(dp.idle) < dp.maxIdle {
		dp.idle = append(dp.idle, sess)
	} else {
		// Pool is full, close the session
		sess.Close()
		delete(dp.createdAt, sess)
	}
}

// cleanExpired removes expired idle connections
func (dp *DevicePool) cleanExpired() {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	now := time.Now()
	validIdle := make([]Session, 0, len(dp.idle))

	for _, sess := range dp.idle {
		if createdAt, exists := dp.createdAt[sess]; exists {
			if now.Sub(createdAt) < dp.idleTime {
				validIdle = append(validIdle, sess)
			} else {
				sess.Close()
				delete(dp.createdAt, sess)
			}
		} else {
			// No creation time, consider it expired
			sess.Close()
		}
	}

	dp.idle = validIdle
}

// closeAll closes all sessions in the pool
func (dp *DevicePool) closeAll() {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	// Close idle sessions
	for _, sess := range dp.idle {
		sess.Close()
	}
	dp.idle = nil

	// Close active sessions
	for sess := range dp.active {
		sess.Close()
	}
	dp.active = make(map[Session]bool)
	dp.createdAt = make(map[Session]time.Time)
}

// Stats returns pool statistics
func (p *ConnectionPool) Stats() map[MinerID]PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := make(map[MinerID]PoolStats)
	for id, pool := range p.pools {
		pool.mu.Lock()
		stats[id] = PoolStats{
			ActiveConnections: len(pool.active),
			IdleConnections:   len(pool.idle),
			MaxOpen:           pool.maxOpen,
			MaxIdle:           pool.maxIdle,
		}
		pool.mu.Unlock()
	}
	return stats
}

// PoolStats contains statistics about a device pool
type PoolStats struct {
	ActiveConnections int `json:"active_connections"`
	IdleConnections   int `json:"idle_connections"`
	MaxOpen           int `json:"max_open"`
	MaxIdle           int `json:"max_idle"`
}
