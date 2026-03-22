package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Theme holds all colors used by the UI.
type Theme struct {
	Name     string
	Primary  color.Color
	Accent   color.Color
	Text     color.Color
	Bright   color.Color
	Muted    color.Color
	Dim      color.Color
	Red      color.Color
	Green    color.Color
	SelectBg color.Color
	SelectFg color.Color
}

var Themes = map[string]Theme{
	"default": {
		Name:     "default",
		Primary:  lipgloss.Color("#7D56F4"),
		Accent:   lipgloss.Color("#7D56F4"),
		Text:     lipgloss.Color("#FAFAFA"),
		Bright:   lipgloss.Color("#FFFFFF"),
		Muted:    lipgloss.Color("#626262"),
		Dim:      lipgloss.Color("#444444"),
		Red:      lipgloss.Color("#FF4444"),
		Green:    lipgloss.Color("#44FF44"),
		SelectBg: lipgloss.Color("#7D56F4"),
		SelectFg: lipgloss.Color("#FFFFFF"),
	},
	"minimal": {
		Name:     "minimal",
		Primary:  lipgloss.Color("#A78BFA"),
		Accent:   lipgloss.Color("#A78BFA"),
		Text:     lipgloss.Color("#E0E0E0"),
		Bright:   lipgloss.Color("#FFFFFF"),
		Muted:    lipgloss.Color("#808080"),
		Dim:      lipgloss.Color("#3A3A3A"),
		Red:      lipgloss.Color("#F87171"),
		Green:    lipgloss.Color("#86EFAC"),
		SelectBg: lipgloss.Color("#333355"),
		SelectFg: lipgloss.Color("#FFFFFF"),
	},
	"ocean": {
		Name:     "ocean",
		Primary:  lipgloss.Color("#22D3EE"),
		Accent:   lipgloss.Color("#38BDF8"),
		Text:     lipgloss.Color("#E0F2FE"),
		Bright:   lipgloss.Color("#FFFFFF"),
		Muted:    lipgloss.Color("#64748B"),
		Dim:      lipgloss.Color("#334155"),
		Red:      lipgloss.Color("#FB7185"),
		Green:    lipgloss.Color("#34D399"),
		SelectBg: lipgloss.Color("#1E3A5F"),
		SelectFg: lipgloss.Color("#FFFFFF"),
	},
	"rose": {
		Name:     "rose",
		Primary:  lipgloss.Color("#F472B6"),
		Accent:   lipgloss.Color("#FB7185"),
		Text:     lipgloss.Color("#FDE8EF"),
		Bright:   lipgloss.Color("#FFFFFF"),
		Muted:    lipgloss.Color("#9D7A8A"),
		Dim:      lipgloss.Color("#4A3040"),
		Red:      lipgloss.Color("#F87171"),
		Green:    lipgloss.Color("#86EFAC"),
		SelectBg: lipgloss.Color("#5E2040"),
		SelectFg: lipgloss.Color("#FFFFFF"),
	},
	"forest": {
		Name:     "forest",
		Primary:  lipgloss.Color("#22C55E"),
		Accent:   lipgloss.Color("#4ADE80"),
		Text:     lipgloss.Color("#DCFCE7"),
		Bright:   lipgloss.Color("#FFFFFF"),
		Muted:    lipgloss.Color("#6B8F71"),
		Dim:      lipgloss.Color("#1A3A2A"),
		Red:      lipgloss.Color("#F87171"),
		Green:    lipgloss.Color("#86EFAC"),
		SelectBg: lipgloss.Color("#14532D"),
		SelectFg: lipgloss.Color("#FFFFFF"),
	},
	"sunset": {
		Name:     "sunset",
		Primary:  lipgloss.Color("#F59E0B"),
		Accent:   lipgloss.Color("#FB923C"),
		Text:     lipgloss.Color("#FEF3C7"),
		Bright:   lipgloss.Color("#FFFFFF"),
		Muted:    lipgloss.Color("#A08050"),
		Dim:      lipgloss.Color("#422006"),
		Red:      lipgloss.Color("#EF4444"),
		Green:    lipgloss.Color("#84CC16"),
		SelectBg: lipgloss.Color("#78350F"),
		SelectFg: lipgloss.Color("#FFFFFF"),
	},
	"nord": {
		Name:     "nord",
		Primary:  lipgloss.Color("#88C0D0"),
		Accent:   lipgloss.Color("#81A1C1"),
		Text:     lipgloss.Color("#ECEFF4"),
		Bright:   lipgloss.Color("#ECEFF4"),
		Muted:    lipgloss.Color("#616E88"),
		Dim:      lipgloss.Color("#3B4252"),
		Red:      lipgloss.Color("#BF616A"),
		Green:    lipgloss.Color("#A3BE8C"),
		SelectBg: lipgloss.Color("#434C5E"),
		SelectFg: lipgloss.Color("#ECEFF4"),
	},
	"dracula": {
		Name:     "dracula",
		Primary:  lipgloss.Color("#BD93F9"),
		Accent:   lipgloss.Color("#FF79C6"),
		Text:     lipgloss.Color("#F8F8F2"),
		Bright:   lipgloss.Color("#F8F8F2"),
		Muted:    lipgloss.Color("#6272A4"),
		Dim:      lipgloss.Color("#44475A"),
		Red:      lipgloss.Color("#FF5555"),
		Green:    lipgloss.Color("#50FA7B"),
		SelectBg: lipgloss.Color("#44475A"),
		SelectFg: lipgloss.Color("#F8F8F2"),
	},
	"solarized": {
		Name:     "solarized",
		Primary:  lipgloss.Color("#268BD2"),
		Accent:   lipgloss.Color("#2AA198"),
		Text:     lipgloss.Color("#93A1A1"),
		Bright:   lipgloss.Color("#FDF6E3"),
		Muted:    lipgloss.Color("#657B83"),
		Dim:      lipgloss.Color("#073642"),
		Red:      lipgloss.Color("#DC322F"),
		Green:    lipgloss.Color("#859900"),
		SelectBg: lipgloss.Color("#073642"),
		SelectFg: lipgloss.Color("#FDF6E3"),
	},
	"monokai": {
		Name:     "monokai",
		Primary:  lipgloss.Color("#F92672"),
		Accent:   lipgloss.Color("#66D9EF"),
		Text:     lipgloss.Color("#F8F8F2"),
		Bright:   lipgloss.Color("#F8F8F2"),
		Muted:    lipgloss.Color("#75715E"),
		Dim:      lipgloss.Color("#3E3D32"),
		Red:      lipgloss.Color("#F92672"),
		Green:    lipgloss.Color("#A6E22E"),
		SelectBg: lipgloss.Color("#49483E"),
		SelectFg: lipgloss.Color("#F8F8F2"),
	},
	"catppuccin": {
		Name:     "catppuccin",
		Primary:  lipgloss.Color("#CBA6F7"),
		Accent:   lipgloss.Color("#F5C2E7"),
		Text:     lipgloss.Color("#CDD6F4"),
		Bright:   lipgloss.Color("#CDD6F4"),
		Muted:    lipgloss.Color("#6C7086"),
		Dim:      lipgloss.Color("#313244"),
		Red:      lipgloss.Color("#F38BA8"),
		Green:    lipgloss.Color("#A6E3A1"),
		SelectBg: lipgloss.Color("#45475A"),
		SelectFg: lipgloss.Color("#CDD6F4"),
	},
}

// ThemeNames returns theme names in display order.
func ThemeNames() []string {
	return []string{"default", "minimal", "ocean", "rose", "forest", "sunset", "nord", "dracula", "solarized", "monokai", "catppuccin"}
}

// ApplyTheme updates all package-level styles to use the given theme.
func ApplyTheme(t Theme) {
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Bright).
		Background(t.Primary).
		Padding(0, 1)

	countStyle = lipgloss.NewStyle().
		Foreground(t.Muted)

	helpStyle = lipgloss.NewStyle().
		Foreground(t.Muted)

	helpKeyStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Primary)

	helpDescStyle = lipgloss.NewStyle().
		Foreground(t.Muted)

	helpSepStyle = lipgloss.NewStyle().
		Foreground(t.Dim)

	labelStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Primary).
		Width(14)

	valueStyle = lipgloss.NewStyle().
		Foreground(t.Text)

	userStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Green)

	assistantStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Accent)

	dimmedStyle = lipgloss.NewStyle().
		Foreground(t.Muted)

	deletePromptStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Red)

	filterStyle = lipgloss.NewStyle().
		Foreground(t.Primary)

	// Store for table styles
	currentTheme = t
}

var currentTheme = Themes["default"]

// CurrentTheme returns the active theme.
func CurrentTheme() Theme {
	return currentTheme
}
