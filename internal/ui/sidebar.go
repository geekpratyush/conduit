package ui

import (
	"fmt"
	"image/color"
	"sort"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/geekpratyush/conduit/internal/plugin"
	ctheme "github.com/geekpratyush/conduit/internal/ui/theme"
)

// buildSidebar constructs the colour-coded connection tree: profiles grouped by
// domain, each row tinted by its protocol's semantic colour.
func (a *App) buildSidebar() *fyne.Container {
	title := widget.NewLabelWithStyle("CONNECTIONS", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	newBtn := widget.NewButton("＋ New Connection", a.showNewConnectionDialog)
	a.connList = container.NewVBox()
	a.populateConnections()
	scroll := container.NewVScroll(a.connList)
	return container.NewBorder(container.NewVBox(title, newBtn, widget.NewSeparator()), nil, nil, nil, scroll)
}

// populateConnections (re)builds the grouped connection rows from the store.
func (a *App) populateConnections() {
	a.connList.Objects = nil
	if a.store == nil {
		a.connList.Add(widget.NewLabel("(no connection store)"))
		a.connList.Refresh()
		return
	}

	// Group profiles by category, preserving a stable category order.
	byCat := map[string][]plugin.ConnectionConfig{}
	for _, c := range a.store.List() {
		cat := categoryForProtocol(c.Protocol)
		byCat[cat] = append(byCat[cat], c)
	}
	cats := make([]string, 0, len(byCat))
	for cat := range byCat {
		cats = append(cats, cat)
	}
	sort.Strings(cats)

	for _, cat := range cats {
		catColor := ctheme.DomainColor(cat)
		a.connList.Add(categoryHeader(cat, catColor))
		for _, c := range byCat[cat] {
			a.connList.Add(a.connectionRow(c, catColor))
		}
	}
	a.connList.Refresh()
}

func categoryHeader(name string, col color.NRGBA) fyne.CanvasObject {
	dot := coloredDot(col)
	lbl := widget.NewLabelWithStyle(name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	return container.NewHBox(dot, lbl)
}

func (a *App) connectionRow(c plugin.ConnectionConfig, col color.NRGBA) fyne.CanvasObject {
	dot := coloredDot(col)
	btn := widget.NewButton(c.Name, func() { a.openConnectionTab(c) })
	btn.Alignment = widget.ButtonAlignLeading
	return container.NewBorder(nil, nil, dot, nil, btn)
}

// coloredDot returns a small filled circle in col, used to tint sidebar rows.
func coloredDot(col color.NRGBA) fyne.CanvasObject {
	c := canvas.NewCircle(col)
	return container.NewGridWrap(fyne.NewSize(12, 12), c)
}

// showNewConnectionDialog collects a new connection and saves it to the store.
func (a *App) showNewConnectionDialog() {
	if a.store == nil {
		dialog.ShowInformation("Unavailable", "No connection store is open.", a.win)
		return
	}
	name := widget.NewEntry()
	name.SetPlaceHolder("My API")

	protocols := []string{}
	for _, d := range a.ctx.Registry.Descriptors() {
		protocols = append(protocols, d.Protocol)
	}
	if len(protocols) == 0 {
		protocols = []string{"rest"}
	}
	proto := widget.NewSelect(protocols, nil)
	proto.SetSelected(protocols[0])

	target := widget.NewEntry()
	target.SetPlaceHolder("https://api.example.com  (or host:port)")

	items := []*widget.FormItem{
		widget.NewFormItem("Name", name),
		widget.NewFormItem("Protocol", proto),
		widget.NewFormItem("URL / Target", target),
	}
	dialog.NewForm("New Connection", "Save", "Cancel", items, func(ok bool) {
		if !ok {
			return
		}
		cfg := plugin.ConnectionConfig{
			ID:       fmt.Sprintf("conn-%d", time.Now().UnixNano()),
			Name:     firstNonEmpty(name.Text, "Untitled"),
			Protocol: proto.Selected,
			URL:      target.Text,
			Auth:     plugin.AuthNone,
		}
		if err := a.store.Save(cfg); err != nil {
			dialog.ShowError(err, a.win)
			return
		}
		a.populateConnections()
		a.appendLog("Saved connection " + cfg.Name)
		a.openConnectionTab(cfg)
	}, a.win).Show()
}

func (a *App) showAbout() {
	msg := "Conduit — a universal protocol workbench.\n\n" +
		"One Console. Every Protocol.\n\n" +
		"Author: Pratyush Ranjan Mishra\n" +
		"github.com/geekpratyush/conduit"
	dialog.ShowInformation("About Conduit", msg, a.win)
}

// categoryForProtocol maps a protocol id to its semantic domain category so
// rows can be coloured even before that protocol's connector is registered.
func categoryForProtocol(p string) string {
	switch p {
	case "rest", "http", "ws", "websocket", "sse", "graphql", "grpc":
		return "HTTP & Web"
	case "kafka", "mqtt", "rabbitmq", "jms", "sqs", "sns":
		return "Messaging"
	case "sql", "postgres", "mysql", "mariadb", "sqlite", "mongo", "mongodb", "redis":
		return "Databases"
	case "sftp", "ftp", "ftps", "s3", "azure", "gcs":
		return "Files & Objects"
	case "ldap", "snmp":
		return "Directory & Monitoring"
	case "mcp", "llm", "agent":
		return "AI"
	default:
		return "HTTP & Web"
	}
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
