// Command conduit is the entry point for the Conduit universal protocol
// workbench: it wires the shared services (event bus, cache, registry,
// environment, history, connection store, credential vault) and launches the
// Fyne desktop shell.
package main

import (
	"fmt"
	"os"

	"github.com/geekpratyush/conduit/internal/core"
	"github.com/geekpratyush/conduit/internal/protocol/httpc"
	"github.com/geekpratyush/conduit/internal/security"
	"github.com/geekpratyush/conduit/internal/ui"
)

func main() {
	app := core.Bootstrap()

	// Register available protocol connectors. More are wired in as their views
	// land (see TASKS.md).
	app.Registry.Register(httpc.Descriptor())

	// History store (best-effort; the app still runs without it).
	if path, err := core.ConfigPath("history.db"); err == nil {
		if hs, err := core.OpenHistory(path); err == nil {
			app.History = hs
			defer hs.Close()
		} else {
			fmt.Fprintf(os.Stderr, "warning: history store unavailable: %v\n", err)
		}
	}

	// Connection-profile store (seeds public samples on first run).
	var store *core.ConnectionStore
	if path, err := core.ConfigPath("connections.json"); err == nil {
		if s, err := core.OpenConnectionStore(path); err == nil {
			store = s
		} else {
			fmt.Fprintf(os.Stderr, "warning: connection store unavailable: %v\n", err)
		}
	}

	// Credential vault (created/unlocked from the status-bar control).
	var vault *security.CredentialVault
	if path, err := core.ConfigPath("vault.enc"); err == nil {
		vault = security.NewVault(path)
	}

	ui.Run(app, store, vault)
}
