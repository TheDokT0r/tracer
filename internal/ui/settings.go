package ui

import (
	"fmt"
	"strings"

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
	case settingShowDate, settingShowDirectory, settingShowBranch, settingConfirmDelete:
		sv.cycleRight() // toggle is the same both ways
	}
}

func (sv settingsView) view() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("settings"))
	b.WriteString("\n\n")

	items := []struct {
		label string
		value string
	}{
		{"Theme", sv.cfg.Theme},
		{"Sort by", sv.cfg.SortBy},
		{"Show date", boolDisplay(sv.cfg.ShowDate)},
		{"Show directory", boolDisplay(sv.cfg.ShowDirectory)},
		{"Show branch", boolDisplay(sv.cfg.ShowBranch)},
		{"Confirm delete", boolDisplay(sv.cfg.ConfirmDelete)},
	}

	for i, item := range items {
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

func boolDisplay(v bool) string {
	if v {
		return "on"
	}
	return "off"
}
