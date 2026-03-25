package ui

import (
	tea "charm.land/bubbletea/v2"
)

func (a App) updateSettings(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		// Handle confirm exit prompt
		if a.settings.confirmExit {
			switch msg.String() {
			case "y", "Y":
				a.settings.save()
				a.applySettings()
				a.view = viewList
			case "n", "N":
				a.view = viewList
			case "esc":
				a.settings.confirmExit = false
			}
			return a, nil
		}

		switch msg.String() {
		case "esc", "q":
			if a.settings.dirty {
				a.settings.confirmExit = true
				return a, nil
			}
			a.view = viewList
			return a, nil
		case "ctrl+s", "super+s":
			a.settings.save()
			a.applySettings()
			return a, nil
		case "up", "k":
			if a.settings.cursor > 0 {
				a.settings.cursor--
			}
			return a, nil
		case "down", "j":
			if a.settings.cursor < len(a.settings.items)-1 {
				a.settings.cursor++
			}
			return a, nil
		case "right", "l", "enter":
			a.settings.cycle(1)
			return a, nil
		case "left", "h":
			a.settings.cycle(-1)
			return a, nil
		case "ctrl+c":
			return a, tea.Quit
		}
	}
	return a, nil
}

// applySettings copies saved config to the app and rebuilds affected views.
func (a *App) applySettings() {
	a.cfg = a.settings.cfg
	a.list.cfg = a.cfg
	a.list.sortSessions()
	a.list.rebuildTable()
}
