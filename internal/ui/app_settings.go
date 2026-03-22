package ui

import (
	tea "charm.land/bubbletea/v2"
	"tracer/internal/config"
)

func (a App) updateSettings(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "q":
			a.cfg = a.settings.cfg
			config.SaveConfig(a.cfg)
			a.list.cfg = a.cfg
			a.list.sortSessions()
			a.list.rebuildTable()
			a.view = viewList
			return a, nil
		case "up", "k":
			if a.settings.cursor > 0 {
				a.settings.cursor--
			}
			return a, nil
		case "down", "j":
			if a.settings.cursor < int(settingCount)-1 {
				a.settings.cursor++
			}
			return a, nil
		case "right", "l", "enter":
			a.settings.cycleRight()
			return a, nil
		case "left", "h":
			a.settings.cycleLeft()
			return a, nil
		case "ctrl+c":
			return a, tea.Quit
		}
	}
	return a, nil
}
