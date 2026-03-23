package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"tracer/internal/config"
)

type settingType int

const (
	settingTheme settingType = iota
	settingSortBy
	settingShowDate
	settingShowDirectory
	settingShowBranch
	settingConfirmDelete
	settingAutoUpdate
	settingCmdDropdown
	settingCmdGhost
	settingCmdMaxSuggestions
	settingCount
)

type settingsView struct {
	cfg    config.Config
	cursor int
	width  int
	height int
}

func newSettingsView(cfg config.Config, width, height int) settingsView {
	return settingsView{
		cfg:    cfg,
		cursor: 0,
		width:  width,
		height: height,
	}
}

func (sv *settingsView) cycleRight() {
	switch settingType(sv.cursor) {
	case settingTheme:
		names := ThemeNames()
		for i, n := range names {
			if n == sv.cfg.Theme {
				sv.cfg.Theme = names[(i+1)%len(names)]
				ApplyTheme(Themes[sv.cfg.Theme])
				break
			}
		}
	case settingSortBy:
		sorts := []string{"date", "name", "directory"}
		for i, s := range sorts {
			if s == sv.cfg.SortBy {
				sv.cfg.SortBy = sorts[(i+1)%len(sorts)]
				break
			}
		}
	case settingShowDate:
		sv.cfg.ShowDate = !sv.cfg.ShowDate
	case settingShowDirectory:
		sv.cfg.ShowDirectory = !sv.cfg.ShowDirectory
	case settingShowBranch:
		sv.cfg.ShowBranch = !sv.cfg.ShowBranch
	case settingConfirmDelete:
		sv.cfg.ConfirmDelete = !sv.cfg.ConfirmDelete
	case settingAutoUpdate:
		sv.cfg.AutoUpdate = !sv.cfg.AutoUpdate
	case settingCmdDropdown:
		sv.cfg.CmdDropdown = !sv.cfg.CmdDropdown
	case settingCmdGhost:
		sv.cfg.CmdGhost = !sv.cfg.CmdGhost
	case settingCmdMaxSuggestions:
		sv.cfg.CmdMaxSuggestions++
		if sv.cfg.CmdMaxSuggestions > 12 {
			sv.cfg.CmdMaxSuggestions = 3
		}
	}
}

func (sv *settingsView) cycleLeft() {
	switch settingType(sv.cursor) {
	case settingTheme:
		names := ThemeNames()
		for i, n := range names {
			if n == sv.cfg.Theme {
				idx := i - 1
				if idx < 0 {
					idx = len(names) - 1
				}
				sv.cfg.Theme = names[idx]
				ApplyTheme(Themes[sv.cfg.Theme])
				break
			}
		}
	case settingSortBy:
		sorts := []string{"date", "name", "directory"}
		for i, s := range sorts {
			if s == sv.cfg.SortBy {
				idx := i - 1
				if idx < 0 {
					idx = len(sorts) - 1
				}
				sv.cfg.SortBy = sorts[idx]
				break
			}
		}
	case settingShowDate, settingShowDirectory, settingShowBranch, settingConfirmDelete, settingAutoUpdate, settingCmdDropdown, settingCmdGhost:
		sv.cycleRight() // toggle is the same both ways
	case settingCmdMaxSuggestions:
		sv.cfg.CmdMaxSuggestions--
		if sv.cfg.CmdMaxSuggestions < 3 {
			sv.cfg.CmdMaxSuggestions = 12
		}
	}
}

func (sv settingsView) view() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("settings"))
	b.WriteString("\n\n")

	type settingItem struct {
		label   string
		value   string
		section string // non-empty = render section header before this item
	}

	items := []settingItem{
		{label: "Theme", value: sv.cfg.Theme},
		{label: "Sort by", value: sv.cfg.SortBy},
		{label: "Show date", value: boolDisplay(sv.cfg.ShowDate)},
		{label: "Show directory", value: boolDisplay(sv.cfg.ShowDirectory)},
		{label: "Show branch", value: boolDisplay(sv.cfg.ShowBranch)},
		{label: "Confirm delete", value: boolDisplay(sv.cfg.ConfirmDelete)},
		{label: "Auto update", value: boolDisplay(sv.cfg.AutoUpdate)},
		{label: "Cmd dropdown", value: boolDisplay(sv.cfg.CmdDropdown), section: "Command Palette"},
		{label: "Ghost suggest", value: boolDisplay(sv.cfg.CmdGhost)},
		{label: "Max suggestions", value: fmt.Sprintf("%d", sv.cfg.CmdMaxSuggestions)},
	}

	for i, item := range items {
		if item.section != "" {
			b.WriteString("\n")
			b.WriteString(dimmedStyle.Render("  ── " + item.section + " ──"))
			b.WriteString("\n")
		}

		cursor := "  "
		if i == sv.cursor {
			cursor = "> "
		}

		label := fmt.Sprintf("%-18s", item.label)

		if i == sv.cursor {
			line := helpKeyStyle.Render(cursor) +
				valueStyle.Render(label) +
				helpKeyStyle.Render("< ") +
				titleStyle.Render(fmt.Sprintf(" %s ", item.value)) +
				helpKeyStyle.Render(" >")
			b.WriteString(line)
		} else {
			line := dimmedStyle.Render(cursor) +
				dimmedStyle.Render(label) +
				dimmedStyle.Render("  ") +
				helpDescStyle.Render(item.value)
			b.WriteString(line)
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	sep := helpSepStyle.Render(" • ")
	b.WriteString(
		helpKeyStyle.Render("↑/↓") + helpDescStyle.Render(" navigate") + sep +
			helpKeyStyle.Render("←/→") + helpDescStyle.Render(" change") + sep +
			helpKeyStyle.Render("esc") + helpDescStyle.Render(" save & back"),
	)

	return b.String()
}

// SettingsApp wraps settingsView as a standalone tea.Model for the subcommand.
type SettingsApp struct {
	sv       settingsView
	saved    bool
}

func NewSettingsApp(cfg config.Config) SettingsApp {
	return SettingsApp{
		sv: newSettingsView(cfg, 80, 24),
	}
}

func (sa SettingsApp) Init() tea.Cmd { return nil }

func (sa SettingsApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		sa.sv.width = msg.Width
		sa.sv.height = msg.Height
		return sa, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "q":
			sa.saved = true
			return sa, tea.Quit
		case "ctrl+c":
			return sa, tea.Quit
		case "up", "k":
			if sa.sv.cursor > 0 {
				sa.sv.cursor--
			}
		case "down", "j":
			if sa.sv.cursor < int(settingCount)-1 {
				sa.sv.cursor++
			}
		case "right", "l", "enter":
			sa.sv.cycleRight()
		case "left", "h":
			sa.sv.cycleLeft()
		}
	}
	return sa, nil
}

func (sa SettingsApp) View() tea.View {
	return tea.NewView(sa.sv.view())
}

// Config returns the (possibly modified) config.
func (sa SettingsApp) Config() config.Config {
	return sa.sv.cfg
}

// Saved returns true if the user exited normally (not ctrl+c without saving intent).
func (sa SettingsApp) Saved() bool {
	return sa.saved
}

func boolDisplay(v bool) string {
	if v {
		return "on"
	}
	return "off"
}
