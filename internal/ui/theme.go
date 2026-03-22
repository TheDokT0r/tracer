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
	},
}

// ThemeNames returns theme names in display order.
func ThemeNames() []string {
	return []string{"default", "minimal", "ocean", "rose"}
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

	pinStyle = lipgloss.NewStyle().
		Foreground(t.Primary)

	dividerStyle = lipgloss.NewStyle().
		Foreground(t.Dim)

	// Store for table styles
	currentTheme = t
}

var currentTheme = Themes["default"]

// CurrentTheme returns the active theme.
func CurrentTheme() Theme {
	return currentTheme
}
