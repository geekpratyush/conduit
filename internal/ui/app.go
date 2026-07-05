// Package ui builds Conduit's Fyne desktop shell — the main window, the
// colour-coded connection sidebar, the tabbed workspace, the log panel, and the
// status bar — on top of the shared services in internal/core and
// internal/security.
package ui

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"github.com/geekpratyush/conduit/internal/core"
	"github.com/geekpratyush/conduit/internal/plugin"
	"github.com/geekpratyush/conduit/internal/security"
	ctheme "github.com/geekpratyush/conduit/internal/ui/theme"
)

// App owns the shell's widgets and shared services.
type App struct {
	fyneApp fyne.App
	win     fyne.Window

	ctx   *core.AppContext
	store *core.ConnectionStore
	vault *security.CredentialVault
	theme *ctheme.Conduit

	tabs     *container.DocTabs
	logBox   *widget.Entry
	status   *widget.Label
	lockBtn  *widget.Button
	sidebar  *fyne.Container
	connList *fyne.Container
}

// Run constructs and shows the Conduit workbench, blocking until the window is
// closed. store and vault may be nil (the shell degrades gracefully).
func Run(ctx *core.AppContext, store *core.ConnectionStore, vault *security.CredentialVault) {
	fa := app.NewWithID("com.geekpratyush.conduit")
	th := ctheme.New(true) // start in Midnight (dark)
	fa.Settings().SetTheme(th)

	a := &App{
		fyneApp: fa,
		ctx:     ctx,
		store:   store,
		vault:   vault,
		theme:   th,
	}
	a.win = fa.NewWindow("Conduit — One Console. Every Protocol.")
	a.win.Resize(fyne.NewSize(1180, 720))

	root := a.buildRoot()
	a.win.SetMainMenu(a.buildMenu())
	a.registerShortcuts()
	a.win.SetContent(root)

	a.appendLog("Conduit ready. " + a.protocolSummary())
	a.refreshLockIndicator()
	a.win.ShowAndRun()
}

// buildRoot constructs the full shell layout (sidebar | tabs-over-log, with the
// status bar pinned to the bottom). It is separated from Run so the same widget
// tree can be rendered headlessly for previews.
func (a *App) buildRoot() fyne.CanvasObject {
	a.buildStatusBar()
	a.buildLog()
	a.buildTabs()
	a.sidebar = a.buildSidebar()

	center := container.NewVSplit(a.tabs, a.logPanel())
	center.SetOffset(0.78)
	body := container.NewHSplit(a.sidebar, center)
	body.SetOffset(0.22)
	return container.NewBorder(nil, a.statusRow(), nil, nil, body)
}

// NewForPreview builds an App wired for headless rendering (no Fyne GL app or
// window). Used by the preview renderer to rasterize the shell to an image.
func NewForPreview(ctx *core.AppContext, store *core.ConnectionStore, dark bool) *App {
	return &App{ctx: ctx, store: store, theme: ctheme.New(dark)}
}

// Root returns the shell's root layout for headless rendering.
func (a *App) Root() fyne.CanvasObject {
	root := a.buildRoot()
	a.appendLog("Conduit ready. " + a.protocolSummary())
	a.refreshLockIndicator()
	return root
}

func (a *App) protocolSummary() string {
	n := len(a.ctx.Registry.Descriptors())
	return fmt.Sprintf("%d protocol(s) registered.", n)
}

// --- Menu & shortcuts ---

func (a *App) buildMenu() *fyne.MainMenu {
	file := fyne.NewMenu("File",
		fyne.NewMenuItem("New Connection…", a.showNewConnectionDialog),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", func() { a.fyneApp.Quit() }),
	)
	toggle := fyne.NewMenuItem("Toggle Dark / Light Theme", a.toggleTheme)
	toggle.Shortcut = &desktop.CustomShortcut{KeyName: fyne.KeyT, Modifier: fyne.KeyModifierControl | fyne.KeyModifierShift}
	view := fyne.NewMenu("View", toggle)
	help := fyne.NewMenu("Help",
		fyne.NewMenuItem("About Conduit", a.showAbout),
	)
	return fyne.NewMainMenu(file, view, help)
}

func (a *App) registerShortcuts() {
	a.win.Canvas().AddShortcut(
		&desktop.CustomShortcut{KeyName: fyne.KeyT, Modifier: fyne.KeyModifierControl | fyne.KeyModifierShift},
		func(fyne.Shortcut) { a.toggleTheme() },
	)
}

func (a *App) toggleTheme() {
	a.theme.Dark = !a.theme.Dark
	a.fyneApp.Settings().SetTheme(a.theme)
	a.updateStatus()
	mode := "Daylight (light)"
	if a.theme.Dark {
		mode = "Midnight (dark)"
	}
	a.appendLog("Theme → " + mode)
}

// --- Status bar ---

func (a *App) buildStatusBar() {
	a.status = widget.NewLabel("")
	a.lockBtn = widget.NewButton("Vault", a.onLockButton)
	a.updateStatus()
}

func (a *App) statusRow() fyne.CanvasObject {
	return container.NewBorder(widget.NewSeparator(), nil, nil, a.lockBtn, a.status)
}

func (a *App) updateStatus() {
	mode := "Daylight"
	if a.theme.Dark {
		mode = "Midnight"
	}
	a.status.SetText(fmt.Sprintf("Theme: %s   •   %s", mode, a.protocolSummary()))
}

// --- Vault lock indicator ---

func (a *App) refreshLockIndicator() {
	if a.vault == nil {
		a.lockBtn.SetText("Vault: n/a")
		a.lockBtn.Disable()
		return
	}
	switch {
	case !a.vault.Exists():
		a.lockBtn.SetText("🔓 Vault: not set up")
	case a.vault.Locked():
		a.lockBtn.SetText("🔒 Vault: locked")
	default:
		a.lockBtn.SetText("🔓 Vault: unlocked")
	}
}

func (a *App) onLockButton() {
	if a.vault == nil {
		return
	}
	if !a.vault.Exists() {
		a.promptVaultPassword("Set a master password", func(pw string) error {
			return a.vault.Initialize(pw)
		})
		return
	}
	if a.vault.Locked() {
		a.promptVaultPassword("Unlock vault", func(pw string) error {
			return a.vault.Unlock(pw)
		})
		return
	}
	a.vault.Lock()
	a.appendLog("Vault locked.")
	a.refreshLockIndicator()
}

func (a *App) promptVaultPassword(title string, action func(string) error) {
	pw := widget.NewPasswordEntry()
	pw.SetPlaceHolder("master password")
	form := dialog.NewForm(title, "OK", "Cancel",
		[]*widget.FormItem{widget.NewFormItem("Password", pw)},
		func(ok bool) {
			if !ok {
				return
			}
			if err := action(pw.Text); err != nil {
				dialog.ShowError(err, a.win)
				a.appendLog("Vault: " + err.Error())
				return
			}
			a.appendLog("Vault: " + title + " ✓")
			a.refreshLockIndicator()
		}, a.win)
	form.Show()
}

// --- Log panel ---

func (a *App) buildLog() {
	a.logBox = widget.NewMultiLineEntry()
	a.logBox.Wrapping = fyne.TextWrapWord
	a.logBox.Disable()
}

func (a *App) logPanel() fyne.CanvasObject {
	header := widget.NewLabelWithStyle("Log", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	return container.NewBorder(header, nil, nil, nil, container.NewScroll(a.logBox))
}

func (a *App) appendLog(line string) {
	ts := time.Now().Format("15:04:05")
	a.logBox.SetText(a.logBox.Text + fmt.Sprintf("[%s] %s\n", ts, line))
	a.logBox.CursorRow = len(a.logBox.Text)
}

// --- Workspace tabs ---

func (a *App) buildTabs() {
	a.tabs = container.NewDocTabs(
		container.NewTabItem("Welcome", a.welcomeTab()),
	)
}

func (a *App) welcomeTab() fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Conduit", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	sub := widget.NewLabelWithStyle("One Console. Every Protocol.", fyne.TextAlignCenter, fyne.TextStyle{Italic: true})
	hint := widget.NewLabel("Pick a connection from the sidebar, or File → New Connection… to begin.")
	return container.NewCenter(container.NewVBox(title, sub, widget.NewSeparator(), hint))
}

// openConnectionTab opens (or focuses) a workspace tab for a saved connection.
func (a *App) openConnectionTab(cfg plugin.ConnectionConfig) {
	item := container.NewTabItem(cfg.Name, a.connectionView(cfg))
	a.tabs.Append(item)
	a.tabs.Select(item)
	a.appendLog("Opened " + cfg.Name + " (" + cfg.Protocol + ")")
}

// connectionView is a first-cut per-connection panel: it shows the target and a
// "Test connection" action that drives the registered connector end-to-end.
func (a *App) connectionView(cfg plugin.ConnectionConfig) fyne.CanvasObject {
	target := cfg.URL
	if target == "" && cfg.Host != "" {
		target = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	}
	info := widget.NewLabel(fmt.Sprintf("Protocol: %s\nTarget:   %s\nAuth:     %s", cfg.Protocol, target, cfg.Auth))
	result := widget.NewLabel("")
	testBtn := widget.NewButton("Test connection", func() {
		conn, err := a.ctx.Registry.New(cfg.Protocol)
		if err != nil {
			result.SetText("No connector for protocol " + cfg.Protocol + " yet.")
			return
		}
		result.SetText("Testing…")
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			res := conn.Test(ctx, cfg)
			_ = conn.Close()
			msg := fmt.Sprintf("%s — %s (%s)", statusWord(res.OK), res.Message, res.Latency.Round(time.Millisecond))
			fyne.Do(func() {
				result.SetText(msg)
				a.appendLog(cfg.Name + ": " + msg)
			})
		}()
	})
	return container.NewBorder(info, nil, nil, nil, container.NewVBox(testBtn, result))
}

func statusWord(ok bool) string {
	if ok {
		return "OK"
	}
	return "FAILED"
}
