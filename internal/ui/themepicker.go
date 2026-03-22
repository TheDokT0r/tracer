package ui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type ThemePicker struct {
	themes   []string
	cursor   int
	current  string
	chosen   string
	table    table.Model
	width    int
	height   int
	quitting bool
}

func NewThemePicker(current string) ThemePicker {
	names := ThemeNames()
	cursor := 0
	for i, n := range names {
		if n == current {
			cursor = i
			break
		}
	}
	tp := ThemePicker{
		themes:  names,
		cursor:  cursor,
		current: current,
		width:   80,
		height:  24,
	}
	tp.applyThemePreview()
	return tp
}

func (tp *ThemePicker) applyThemePreview() {
	name := tp.themes[tp.cursor]
	t := Themes[name]
	ApplyTheme(t)
	tp.rebuildTable(t)
}

func (tp *ThemePicker) rebuildTable(t Theme) {
	cols := []table.Column{
		{Title: "Name", Width: 30},
		{Title: "Date", Width: 18},
		{Title: "Directory", Width: 20},
		{Title: "Branch", Width: 20},
	}
	rows := []table.Row{
		{"* Add auth middleware", "2026-03-22 14:30", "~/projects/my-app", "main"},
		{"Fix login bug", "2026-03-21 09:15", "~/projects/api", "feat/login"},
		{"Refactor database", "2026-03-20 16:45", "~/projects/api", "develop"},
		{"Create CLI tool", "2026-03-20 11:00", "~/projects", "-"},
		{"Write unit tests", "2026-03-19 20:30", "~/projects/lib", "main"},
	}

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

	tp.table = table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(5),
	)
	tp.table.SetStyles(s)
}

func (tp ThemePicker) Init() tea.Cmd {
	return nil
}

func (tp ThemePicker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		tp.width = msg.Width
		tp.height = msg.Height
		return tp, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			// Restore original theme
			ApplyTheme(Themes[tp.current])
			tp.quitting = true
			return tp, tea.Quit
		case "enter":
			tp.chosen = tp.themes[tp.cursor]
			tp.quitting = true
			return tp, tea.Quit
		case "left", "h":
			if tp.cursor > 0 {
				tp.cursor--
				tp.applyThemePreview()
			}
			return tp, nil
		case "right", "l":
			if tp.cursor < len(tp.themes)-1 {
				tp.cursor++
				tp.applyThemePreview()
			}
			return tp, nil
		}
	}
	return tp, nil
}

func (tp ThemePicker) View() tea.View {
	var b strings.Builder

	b.WriteString(titleStyle.Render("tracer"))
	b.WriteString(countStyle.Render("  theme picker"))
	b.WriteString("\n\n")

	// Theme tabs
	for i, name := range tp.themes {
		if i == tp.cursor {
			b.WriteString(helpKeyStyle.Render(fmt.Sprintf(" [%s] ", name)))
		} else {
			b.WriteString(helpDescStyle.Render(fmt.Sprintf("  %s  ", name)))
		}
	}
	b.WriteString("\n\n")

	// Preview table
	b.WriteString(tp.table.View())
	b.WriteString("\n\n")

	// Sample detail preview
	b.WriteString(labelStyle.Render("Session ID") + valueStyle.Render("abc-123-def-456") + "\n")
	b.WriteString(labelStyle.Render("Directory") + valueStyle.Render("~/projects/my-app") + "\n")
	b.WriteString(labelStyle.Render("Branch") + valueStyle.Render("main") + "\n")
	b.WriteString("\n")

	// Sample conversation
	b.WriteString(userStyle.Render("You: ") + "Add auth middleware to the API\n")
	b.WriteString(assistantStyle.Render("Claude: ") + "I'll add JWT authentication middleware...\n")
	b.WriteString("\n")

	// Help
	sep := helpSepStyle.Render(" • ")
	b.WriteString(
		helpKeyStyle.Render("←/→") + helpDescStyle.Render(" switch theme") + sep +
			helpKeyStyle.Render("enter") + helpDescStyle.Render(" apply") + sep +
			helpKeyStyle.Render("esc") + helpDescStyle.Render(" cancel"),
	)

	return tea.NewView(b.String())
}

// Chosen returns the selected theme name, or empty if cancelled.
func (tp ThemePicker) Chosen() string {
	return tp.chosen
}
