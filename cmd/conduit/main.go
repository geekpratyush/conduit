// Command conduit is the entry point for the Conduit universal protocol
// workbench.
//
// This is currently a bootstrap launcher: it constructs the shared AppContext
// (event bus, cache, protocol registry, environment service), opens the history
// store, and reports what is wired. The Fyne desktop shell (window, sidebar
// connection tree, tabbed workspace) is layered on top of this same AppContext
// in the next phase — see TASKS.md.
package main

import (
	"fmt"
	"os"

	"github.com/geekpratyush/conduit/internal/core"
	"github.com/geekpratyush/conduit/internal/protocol/httpc"
)

func main() {
	fmt.Println("Conduit — One Console. Every Protocol.")

	app := core.Bootstrap()

	// Register available protocol connectors. Each protocol package contributes
	// a Descriptor; the shell lists these in the sidebar and builds connectors
	// on demand. More are wired in as their views land (see TASKS.md).
	app.Registry.Register(httpc.Descriptor())

	// Attach the history store; a failure here must not stop the app.
	if path, err := core.ConfigPath("history.db"); err == nil {
		if hs, err := core.OpenHistory(path); err == nil {
			app.History = hs
			defer hs.Close()
		} else {
			fmt.Fprintf(os.Stderr, "warning: history store unavailable: %v\n", err)
		}
	}

	protocols := app.Registry.Descriptors()
	fmt.Printf("Core services ready: eventbus, cache, env, registry (%d protocols registered).\n", len(protocols))
	if app.History != nil {
		fmt.Println("History store: open (SQLite + FTS5).")
	}
	fmt.Println("Desktop shell (Fyne) not yet wired — run `go test ./internal/...` to exercise the core.")
}
