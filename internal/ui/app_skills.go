package ui

import (
	"os"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	skillspkg "tracer/internal/skills"
)

func (a App) updateSkillsList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if a.skillsList.filtering {
			switch msg.String() {
			case "esc":
				a.skillsList.filtering = false
				a.skillsList.filter.SetValue("")
				a.skillsList.applyFilter()
				return a, nil
			case "enter":
				a.skillsList.filtering = false
				return a, nil
			default:
				var cmd tea.Cmd
				a.skillsList.filter, cmd = a.skillsList.filter.Update(msg)
				a.skillsList.applyFilter()
				return a, cmd
			}
		}

		if a.newSkill {
			return a.updateNewSkill(msg)
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit
		case "tab":
			return a.nextTab()
		case "shift+tab":
			return a.prevTab()
		case "/":
			a.skillsList.filtering = true
			a.skillsList.filter.Focus()
			return a, nil
		case "enter", "v":
			return a.openSkillDetail()
		case "e":
			return a.editSkillFile()
		case "n":
			return a.startNewSkill()
		case "d":
			if sk := a.skillsList.selectedSkill(); sk != nil && !sk.ReadOnly {
				if a.cfg.ConfirmDelete {
					a.confirmDelete = true
				} else {
					skillspkg.DeleteSkill(*sk)
					a.skillsList.removeSkill(sk.Name)
				}
			}
			return a, nil
		case ":":
			if !a.anyModalActive() {
				return a.enterCommandMode()
			}
		}
	}

	var cmd tea.Cmd
	a.skillsList.table, cmd = a.skillsList.table.Update(msg)
	return a, cmd
}

func (a App) updateSkillDetail(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "q":
			a.view = viewSkillsList
			return a, nil
		case "e":
			return a.editSkillFile()
		case "d":
			if sk := a.currentSkill(); sk != nil && !sk.ReadOnly {
				if a.cfg.ConfirmDelete {
					a.confirmDelete = true
				} else {
					skillspkg.DeleteSkill(*sk)
					a.skillsList.removeSkill(sk.Name)
					a.view = viewSkillsList
				}
			}
			return a, nil
		case ":":
			return a.enterCommandMode()
		case "ctrl+c":
			return a, tea.Quit
		}
	}

	var cmd tea.Cmd
	a.skillDetail.viewport, cmd = a.skillDetail.viewport.Update(msg)
	return a, cmd
}

// --- Skill actions ---

func (a App) openSkillDetail() (tea.Model, tea.Cmd) {
	sk := a.skillsList.selectedSkill()
	if sk == nil {
		return a, nil
	}
	content, err := os.ReadFile(sk.Path)
	if err != nil {
		return a, nil
	}
	a.skillDetail = newSkillDetailView(*sk, string(content), a.width, a.height)
	a.view = viewSkillDetail
	return a, nil
}

func (a App) currentSkill() *skillspkg.Skill {
	if a.view == viewSkillDetail {
		sk := a.skillDetail.skill
		return &sk
	}
	return a.skillsList.selectedSkill()
}

func (a App) editSkillFile() (tea.Model, tea.Cmd) {
	sk := a.currentSkill()
	if sk == nil || sk.ReadOnly {
		return a, nil
	}
	return a, openEditor(sk.Path)
}

func (a App) createSkillDirect(name string) tea.Cmd {
	path, err := skillspkg.CreateSkill(a.claudeDir, name, "")
	if err != nil {
		return nil
	}
	return openEditor(path)
}

func (a App) startNewSkill() (tea.Model, tea.Cmd) {
	ti := textinput.New()
	ti.Placeholder = "skill-name"
	ti.Focus()
	ti.CharLimit = 80
	a.newSkillInput = ti
	a.newSkill = true
	return a, nil
}

func (a App) updateNewSkill(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		a.newSkill = false
		return a, nil
	case "enter":
		name := strings.TrimSpace(a.newSkillInput.Value())
		if name == "" {
			a.newSkill = false
			return a, nil
		}
		a.newSkill = false
		path, err := skillspkg.CreateSkill(a.claudeDir, name, "")
		if err != nil {
			return a, nil
		}
		return a, openEditor(path)
	}
	var cmd tea.Cmd
	a.newSkillInput, cmd = a.newSkillInput.Update(msg)
	return a, cmd
}
