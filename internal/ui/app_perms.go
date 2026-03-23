package ui

import (
	tea "charm.land/bubbletea/v2"
	"tracer/internal/ccsettings"
)

func (a App) updatePermsList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if a.permsList.filtering {
			switch msg.String() {
			case "esc":
				a.permsList.filtering = false
				a.permsList.filter.SetValue("")
				a.permsList.applyFilter()
				return a, nil
			case "enter":
				a.permsList.filtering = false
				return a, nil
			default:
				var cmd tea.Cmd
				a.permsList.filter, cmd = a.permsList.filter.Update(msg)
				a.permsList.applyFilter()
				return a, cmd
			}
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit
		case "tab":
			return a.nextTab()
		case "shift+tab":
			return a.prevTab()
		case "/":
			a.permsList.filtering = true
			a.permsList.filter.Focus()
			return a, nil
		case "enter", "v":
			if f := a.permsList.selectedFile(); f != nil {
				a.permsDetail = newPermsDetailView(f, a.width, a.height)
				a.view = viewPermsDetail
			}
			return a, nil
		case ":":
			if !a.anyModalActive() {
				return a.enterCommandMode()
			}
		}
	}

	var cmd tea.Cmd
	a.permsList.table, cmd = a.permsList.table.Update(msg)
	return a, cmd
}

func (a App) updatePermsDetail(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if a.addRule.active {
			done, result := a.addRule.update(msg)
			if done && result != nil {
				a.permsDetail.addRule(result.list, result.rule)
				a.permsList.rebuildTable()
			}
			return a, nil
		}

		switch msg.String() {
		case "esc", "q":
			a.permsList.rebuildTable()
			a.view = viewPermsList
			return a, nil
		case "a":
			a.addRule = newAddRuleState([]ccsettings.SettingsFile{*a.permsDetail.file})
			return a, nil
		case "t":
			a.permsDetail.toggleSelected()
			return a, nil
		case "d":
			a.permsDetail.deleteSelected()
			return a, nil
		case ":":
			if !a.addRule.active {
				return a.enterCommandMode()
			}
		case "ctrl+c":
			return a, tea.Quit
		}
	}

	var cmd tea.Cmd
	a.permsDetail.table, cmd = a.permsDetail.table.Update(msg)
	return a, cmd
}
