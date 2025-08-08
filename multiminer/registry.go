package multiminer

import (
	"context"
	"sync"
)

// Registry stores available drivers and resolves them for endpoints.
type Registry struct {
	mu      sync.RWMutex
	drivers []Driver
}

func NewRegistry() *Registry { return &Registry{} }

func (r *Registry) Register(d Driver) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.drivers = append(r.drivers, d)
}

// Detect finds the first driver that claims the endpoint.
func (r *Registry) Detect(ctx context.Context, ep Endpoint) (Driver, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, d := range r.drivers {
		ok, err := d.Detect(ctx, ep)
		if err != nil {
			return nil, err
		}
		if ok {
			return d, nil
		}
	}
	return nil, NewDriverNotFoundError()
}

// Get returns a driver by its Name().
func (r *Registry) Get(name string) Driver {
	r.mu.RLock(); defer r.mu.RUnlock()
	for _, d := range r.drivers { if d.Name() == name { return d } }
	return nil
}
