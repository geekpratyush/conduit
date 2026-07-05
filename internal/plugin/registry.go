package plugin

import (
	"fmt"
	"sort"
	"sync"
)

// Descriptor is the registration record for a protocol: enough metadata for the
// shell to list it in menus/sidebars and to construct a fresh connector on
// demand. Conduit favours explicit registration over reflection-based service
// loading, so the wiring is greppable.
type Descriptor struct {
	Protocol    string
	DisplayName string
	Category    string // e.g. "HTTP", "Messaging", "Database", "Files", "AI"
	// New builds a fresh, unconnected connector instance.
	New func() Connector
}

// Registry holds the set of known protocol descriptors. It is safe for
// concurrent use.
type Registry struct {
	mu    sync.RWMutex
	items map[string]Descriptor
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{items: make(map[string]Descriptor)}
}

// Register adds d. It panics on a duplicate protocol id or a nil New func,
// since both indicate a programming error at wiring time.
func (r *Registry) Register(d Descriptor) {
	if d.Protocol == "" {
		panic("plugin: Register with empty Protocol")
	}
	if d.New == nil {
		panic(fmt.Sprintf("plugin: Register %q with nil New", d.Protocol))
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, dup := r.items[d.Protocol]; dup {
		panic(fmt.Sprintf("plugin: duplicate protocol %q", d.Protocol))
	}
	r.items[d.Protocol] = d
}

// Lookup returns the descriptor for protocol and whether it was found.
func (r *Registry) Lookup(protocol string) (Descriptor, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.items[protocol]
	return d, ok
}

// New constructs a fresh connector for protocol, or an error if unregistered.
func (r *Registry) New(protocol string) (Connector, error) {
	d, ok := r.Lookup(protocol)
	if !ok {
		return nil, fmt.Errorf("plugin: no connector registered for %q", protocol)
	}
	return d.New(), nil
}

// Descriptors returns all registered descriptors sorted by Category then
// DisplayName — a stable order for menu/sidebar rendering.
func (r *Registry) Descriptors() []Descriptor {
	r.mu.RLock()
	out := make([]Descriptor, 0, len(r.items))
	for _, d := range r.items {
		out = append(out, d)
	}
	r.mu.RUnlock()
	sort.Slice(out, func(i, j int) bool {
		if out[i].Category != out[j].Category {
			return out[i].Category < out[j].Category
		}
		return out[i].DisplayName < out[j].DisplayName
	})
	return out
}
