package core

import "github.com/geekpratyush/conduit/internal/plugin"

// AppContext is the hand-rolled dependency container wiring together the shared
// services every view and connector needs. Conduit favours explicit
// construction over a DI framework, so the wiring is greppable and the startup
// order is obvious.
//
// Fields are populated by Bootstrap and then treated as read-only for the life
// of the process. Services that carry their own locks (EventBus, Cache) are
// safe for concurrent use.
type AppContext struct {
	Events   *EventBus
	Cache    *Cache
	Registry *plugin.Registry
	Env      *EnvironmentService
	// History is populated by Bootstrap once the SQLite store is added
	// (Phase 1); left as an interface-free pointer here to avoid a cyclic or
	// premature dependency. Nil until then.
	History *HistoryStore
}

// Bootstrap constructs the always-available, dependency-light services (event
// bus, cache, protocol registry, env service). Heavier services that need I/O
// (history DB, vault) are attached by their own initialisers so that a failure
// to open, say, the history database does not prevent the app from starting.
func Bootstrap() *AppContext {
	return &AppContext{
		Events:   NewEventBus(),
		Cache:    NewCache(),
		Registry: plugin.NewRegistry(),
		Env:      NewEnvironmentService(),
	}
}
