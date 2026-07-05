package core

import (
	"container/list"
	"sync"
)

// Cache is a set of named, individually-bounded LRU regions pre-registered for
// hot lookups (DNS, TLS, schema, lag, help, …). Implemented on the stdlib
// (container/list) to stay dependency-free; a region evicts its
// least-recently-used entry once it exceeds its capacity.
type Cache struct {
	mu      sync.Mutex
	regions map[string]*region
}

// Well-known region names, pre-registered at construction with sensible caps.
const (
	RegionDNS    = "dns"
	RegionTLS    = "tls"
	RegionSchema = "schema"
	RegionLag    = "lag"
	RegionHelp   = "help"
)

// NewCache returns a Cache with the standard regions pre-registered.
func NewCache() *Cache {
	c := &Cache{regions: make(map[string]*region)}
	for _, r := range []struct {
		name string
		cap  int
	}{
		{RegionDNS, 512},
		{RegionTLS, 256},
		{RegionSchema, 128},
		{RegionLag, 256},
		{RegionHelp, 256},
	} {
		c.regions[r.name] = newRegion(r.cap)
	}
	return c
}

// Region returns the named region, creating it with the default capacity if it
// was not pre-registered.
func (c *Cache) Region(name string) *region {
	c.mu.Lock()
	defer c.mu.Unlock()
	r, ok := c.regions[name]
	if !ok {
		r = newRegion(256)
		c.regions[name] = r
	}
	return r
}

type entry struct {
	key string
	val any
}

// region is a single bounded LRU map, safe for concurrent use.
type region struct {
	mu       sync.Mutex
	capacity int
	ll       *list.List               // front = most recently used
	items    map[string]*list.Element // key -> element holding *entry
}

func newRegion(capacity int) *region {
	return &region{
		capacity: capacity,
		ll:       list.New(),
		items:    make(map[string]*list.Element),
	}
}

// Get returns the value for key and whether it was present, marking it as most
// recently used on a hit.
func (r *region) Get(key string) (any, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if el, ok := r.items[key]; ok {
		r.ll.MoveToFront(el)
		return el.Value.(*entry).val, true
	}
	return nil, false
}

// Put inserts or updates key, evicting the least-recently-used entry if the
// region is over capacity.
func (r *region) Put(key string, val any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if el, ok := r.items[key]; ok {
		r.ll.MoveToFront(el)
		el.Value.(*entry).val = val
		return
	}
	el := r.ll.PushFront(&entry{key: key, val: val})
	r.items[key] = el
	if r.ll.Len() > r.capacity {
		oldest := r.ll.Back()
		if oldest != nil {
			r.ll.Remove(oldest)
			delete(r.items, oldest.Value.(*entry).key)
		}
	}
}

// Len reports the number of entries currently held.
func (r *region) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.ll.Len()
}
