package ui

import (
	"runtime"

	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
)

// metaKey returns "⌘" on macOS and "ctrl" on other platforms.
func metaKey() string {
	if runtime.GOOS == "darwin" {
		return "⌘"
	}
	return "ctrl"
}

var (
	// Colors used directly in command_input.go dropdown rendering.
	purple = lipgloss.Color("#7D56F4")
	white  = lipgloss.Color("#FAFAFA")

	titleStyle        lipgloss.Style
	countStyle        lipgloss.Style
	helpStyle         lipgloss.Style
	labelStyle        lipgloss.Style
	valueStyle        lipgloss.Style
	userStyle         lipgloss.Style
	assistantStyle    lipgloss.Style
	dimmedStyle       lipgloss.Style
	deletePromptStyle lipgloss.Style
	filterStyle       lipgloss.Style
	helpKeyStyle      lipgloss.Style
	helpDescStyle     lipgloss.Style
	helpSepStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#444444"))
)

// helpItem renders a single key + description for the help bar.
func helpItem(key, desc string) string {
	return helpKeyStyle.Render(key) + helpDescStyle.Render(" "+desc)
}

// themedTableStyles returns table styles using the current theme.
func themedTableStyles() table.Styles {
	t := CurrentTheme()
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		Bold(true).
		Foreground(t.Primary)
	s.Selected = s.Selected.
		Foreground(t.SelectFg).
		Background(t.SelectBg).
		Bold(true)
	return s
}
