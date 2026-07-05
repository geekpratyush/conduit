// Command preview renders the Conduit shell to PNG images using Fyne's software
// renderer — no display, no OpenGL, no compositor. It rasterizes the exact same
// widget tree and theme the live app uses, so it is a faithful preview of the
// UI. Usage: go run ./cmd/preview [outdir]
package main

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"github.com/geekpratyush/conduit/internal/core"
	"github.com/geekpratyush/conduit/internal/protocol/httpc"
	"github.com/geekpratyush/conduit/internal/ui"
	ctheme "github.com/geekpratyush/conduit/internal/ui/theme"
)

func main() {
	outDir := "."
	if len(os.Args) > 1 {
		outDir = os.Args[1]
	}
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fatal(err)
	}

	for _, mode := range []struct {
		name string
		dark bool
	}{{"midnight-dark", true}, {"daylight-light", false}} {
		if err := render(outDir, mode.name, mode.dark); err != nil {
			fatal(err)
		}
		fmt.Printf("wrote %s\n", filepath.Join(outDir, "conduit-"+mode.name+".png"))
	}
}

func render(outDir, name string, dark bool) error {
	// A fresh test app per render so the theme applies cleanly to the software canvas.
	app := test.NewApp()
	app.Settings().SetTheme(ctheme.New(dark))

	ctx := core.Bootstrap()
	ctx.Registry.Register(httpc.Descriptor())

	// A throwaway store seeded with the public samples so the sidebar is populated.
	storePath := filepath.Join(os.TempDir(), "conduit-preview-"+name+".json")
	_ = os.Remove(storePath)
	store, err := core.OpenConnectionStore(storePath)
	if err != nil {
		return err
	}

	shell := ui.NewForPreview(ctx, store, dark)
	root := shell.Root()

	w := test.NewWindow(root)
	defer w.Close()
	w.Resize(fyne.NewSize(1200, 760))

	img := w.Canvas().Capture()
	f, err := os.Create(filepath.Join(outDir, "conduit-"+name+".png"))
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "preview:", err)
	os.Exit(1)
}
