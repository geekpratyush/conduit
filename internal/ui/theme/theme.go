// Package theme implements Conduit's Fyne theme: the Midnight (dark) and
// Daylight (light) neutral palettes plus the fixed semantic domain colours
// (Signal Blue, Stream Amber, …) documented in THEME.md. Colour carries
// meaning throughout the UI, so these values are the single source of truth.
package theme

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	ftheme "fyne.io/fyne/v2/theme"
)

// hex parses "#RRGGBB" into an opaque color.NRGBA. It panics on malformed input
// because every colour here is a compile-time constant string.
func hex(s string) color.NRGBA {
	if len(s) != 7 || s[0] != '#' {
		panic("theme: bad hex colour " + s)
	}
	var r, g, b uint8
	if _, err := fmt.Sscanf(s[1:], "%02x%02x%02x", &r, &g, &b); err != nil {
		panic("theme: bad hex colour " + s + ": " + err.Error())
	}
	return color.NRGBA{R: r, G: g, B: b, A: 0xff}
}

// Semantic domain colours — identical in both themes so a Kafka node is always
// amber and a database always emerald. Consumed by the sidebar, tab accent
// bars, and status pills (not part of the fyne.Theme colour set).
var (
	SignalBlue    = hex("#3B82F6") // HTTP & Web
	StreamAmber   = hex("#F59E0B") // Messaging
	DataEmerald   = hex("#10B981") // Databases
	TransferViolet = hex("#8B5CF6") // Files & Objects
	BeaconTeal    = hex("#14B8A6") // Directory & Monitoring
	NeuralMagenta = hex("#EC4899") // AI & MCP
	CipherGold    = hex("#EAB308") // Security

	StatusSuccess = hex("#22C55E")
	StatusWarning = hex("#F59E0B")
	StatusError   = hex("#F43F5E")
	StatusInfo    = hex("#38BDF8")

	BrandIndigo = hex("#6366F1")
	BrandCyan   = hex("#22D3EE")
)

// DomainColor maps a protocol category (as used by plugin.Descriptor.Category)
// to its semantic colour. Unknown categories fall back to the brand indigo.
func DomainColor(category string) color.NRGBA {
	switch category {
	case "HTTP & Web":
		return SignalBlue
	case "Messaging":
		return StreamAmber
	case "Databases":
		return DataEmerald
	case "Files & Objects":
		return TransferViolet
	case "Directory & Monitoring":
		return BeaconTeal
	case "AI":
		return NeuralMagenta
	case "Security":
		return CipherGold
	default:
		return BrandIndigo
	}
}

// midnight (dark) neutral palette.
var midnight = map[fyne.ThemeColorName]color.Color{
	ftheme.ColorNameBackground:        hex("#0D1117"),
	ftheme.ColorNameForeground:        hex("#E6EDF3"),
	ftheme.ColorNameForegroundOnPrimary: hex("#0D1117"),
	ftheme.ColorNamePrimary:           BrandIndigo,
	ftheme.ColorNameButton:            hex("#1F2630"),
	ftheme.ColorNameInputBackground:   hex("#161B22"),
	ftheme.ColorNameInputBorder:       hex("#2D3540"),
	ftheme.ColorNameMenuBackground:    hex("#161B22"),
	ftheme.ColorNameOverlayBackground: hex("#161B22"),
	ftheme.ColorNameSeparator:         hex("#2D3540"),
	ftheme.ColorNamePlaceHolder:       hex("#8B98A9"),
	ftheme.ColorNameScrollBar:         hex("#2D3540"),
	ftheme.ColorNameSelection:         hex("#26324A"),
	ftheme.ColorNameHover:             hex("#1F2630"),
	ftheme.ColorNamePressed:           hex("#26324A"),
	ftheme.ColorNameDisabled:          hex("#5A6675"),
	ftheme.ColorNameError:             StatusError,
	ftheme.ColorNameSuccess:           StatusSuccess,
	ftheme.ColorNameWarning:           StatusWarning,
	ftheme.ColorNameFocus:             BrandCyan,
}

// daylight (light) neutral palette.
var daylight = map[fyne.ThemeColorName]color.Color{
	ftheme.ColorNameBackground:        hex("#F7F9FC"),
	ftheme.ColorNameForeground:        hex("#1A2230"),
	ftheme.ColorNameForegroundOnPrimary: hex("#FFFFFF"),
	ftheme.ColorNamePrimary:           hex("#5457E6"),
	ftheme.ColorNameButton:            hex("#EEF2F7"),
	ftheme.ColorNameInputBackground:   hex("#FFFFFF"),
	ftheme.ColorNameInputBorder:       hex("#D8DFE8"),
	ftheme.ColorNameMenuBackground:    hex("#FFFFFF"),
	ftheme.ColorNameOverlayBackground: hex("#FFFFFF"),
	ftheme.ColorNameSeparator:         hex("#D8DFE8"),
	ftheme.ColorNamePlaceHolder:       hex("#5A6675"),
	ftheme.ColorNameScrollBar:         hex("#D8DFE8"),
	ftheme.ColorNameSelection:         hex("#DCE6FA"),
	ftheme.ColorNameHover:             hex("#EEF2F7"),
	ftheme.ColorNamePressed:           hex("#DCE6FA"),
	ftheme.ColorNameDisabled:          hex("#AAB4C0"),
	ftheme.ColorNameError:             StatusError,
	ftheme.ColorNameSuccess:           StatusSuccess,
	ftheme.ColorNameWarning:           StatusWarning,
	ftheme.ColorNameFocus:             hex("#0EA5C4"),
}

// Conduit is the app theme. It forces its own dark/light choice (via Dark)
// rather than following the OS variant, so the in-app toggle is authoritative.
type Conduit struct {
	Dark bool
}

// compile-time check.
var _ fyne.Theme = (*Conduit)(nil)

// New returns a Conduit theme in the given mode.
func New(dark bool) *Conduit { return &Conduit{Dark: dark} }

func (t *Conduit) palette() map[fyne.ThemeColorName]color.Color {
	if t.Dark {
		return midnight
	}
	return daylight
}

func (t *Conduit) variant() fyne.ThemeVariant {
	if t.Dark {
		return ftheme.VariantDark
	}
	return ftheme.VariantLight
}

// Color returns a palette colour, falling back to the default theme for any
// name we don't override.
func (t *Conduit) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	if c, ok := t.palette()[name]; ok {
		return c
	}
	return ftheme.DefaultTheme().Color(name, t.variant())
}

func (t *Conduit) Font(style fyne.TextStyle) fyne.Resource { return ftheme.DefaultTheme().Font(style) }
func (t *Conduit) Icon(name fyne.ThemeIconName) fyne.Resource {
	return ftheme.DefaultTheme().Icon(name)
}

func (t *Conduit) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case ftheme.SizeNamePadding:
		return 6
	case ftheme.SizeNameInnerPadding:
		return 8
	default:
		return ftheme.DefaultTheme().Size(name)
	}
}
