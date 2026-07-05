// Package core holds Conduit's shared, protocol-agnostic infrastructure: the
// event bus, the AppContext dependency container, the LRU cache, the history
// store, and the environment-variable service. Every protocol view and
// connector reuses these services.
package core

import "sync"

// Event is any value published on the bus. Subscribers filter by concrete type
// via a type switch. Keeping this an empty interface (rather than a fixed enum)
// lets protocols publish their own event structs without modifying core.
type Event any

// Handler receives a published event.
type Handler func(Event)

// Subscription identifies a handler registration so it can be cancelled.
type Subscription struct {
	id  uint64
	bus *EventBus
}

// Unsubscribe removes the handler. It is safe to call more than once.
func (s Subscription) Unsubscribe() {
	if s.bus == nil {
		return
	}
	s.bus.remove(s.id)
}

// EventBus is a simple synchronous typed publish/subscribe hub, safe for
// concurrent use. Handlers run on the goroutine that calls Publish; handlers
// that must not block the publisher should offload to their own goroutine.
type EventBus struct {
	mu       sync.RWMutex
	nextID   uint64
	handlers map[uint64]Handler
}

// NewEventBus returns a ready-to-use bus.
func NewEventBus() *EventBus {
	return &EventBus{handlers: make(map[uint64]Handler)}
}

// Subscribe registers h for all events and returns a Subscription to cancel it.
func (b *EventBus) Subscribe(h Handler) Subscription {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.nextID++
	id := b.nextID
	b.handlers[id] = h
	return Subscription{id: id, bus: b}
}

// Publish delivers e to every current subscriber synchronously.
func (b *EventBus) Publish(e Event) {
	b.mu.RLock()
	hs := make([]Handler, 0, len(b.handlers))
	for _, h := range b.handlers {
		hs = append(hs, h)
	}
	b.mu.RUnlock()
	for _, h := range hs {
		h(e)
	}
}

func (b *EventBus) remove(id uint64) {
	b.mu.Lock()
	delete(b.handlers, id)
	b.mu.Unlock()
}
