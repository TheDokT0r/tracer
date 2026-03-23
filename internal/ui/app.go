package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"tracer/internal/ccsettings"
	"tracer/internal/claude"
	"tracer/internal/config"
	"tracer/internal/model"
	skillspkg "tracer/internal/skills"
)

type viewState int

const (
	viewList viewState = iota
	viewDetail
	viewSettings
	viewSkillsList
	viewSkillDetail
	viewPermsList
	viewPermsDetail
)

type App struct {
	claudeDir     string
	cfg           config.Config
	tab           tabBar
	view          viewState
	list          listView
	detail        detailView
	settings      settingsView
	skillsList    skillsListView
	skillDetail   skillDetailView
	permsList     permsListView
	permsDetail   permsDetailView
	addRule       addRuleState
	renames       map[string]string
	renaming      bool
	renameInput   textinput.Model
	newSession    bool
	newSessionDir textinput.Model
	newSkill      bool
	newSkillInput textinput.Model
	confirmDelete bool
	exportPicker  bool
	statusMsg     string
	width         int
	height        int
}

func NewApp(claudeDir string, sessions []model.Session, pins map[string]bool, cfg config.Config, renames map[string]string, skills []skillspkg.Skill, settingsFiles []ccsettings.SettingsFile) App {
	return App{
		claudeDir:  claudeDir,
		cfg:        cfg,
		renames:    renames,
		tab:        tabBar{active: TabSessions},
		list:       newListView(sessions, pins, cfg, 80, 24),
		skillsList: newSkillsListView(skills, 80, 24),
		permsList:  newPermsListView(settingsFiles, 80, 24),
		view:       viewList,
	}
}

func (a App) Init() tea.Cmd { return nil }

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case editorFinishedMsg:
		return a.handleEditorFinished()
	case statusClearMsg:
		a.statusMsg = ""
		return a, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return a.handleResize(msg)
	}

	if a.confirmDelete {
		return a.updateDeleteConfirm(msg)
	}

	switch a.view {
	case viewList:
		return a.updateSessionList(msg)
	case viewDetail:
		return a.updateSessionDetail(msg)
	case viewSettings:
		return a.updateSettings(msg)
	case viewSkillsList:
		return a.updateSkillsList(msg)
	case viewSkillDetail:
		return a.updateSkillDetail(msg)
	case viewPermsList:
		return a.updatePermsList(msg)
	case viewPermsDetail:
		return a.updatePermsDetail(msg)
	}
	return a, nil
}

func (a App) handleEditorFinished() (tea.Model, tea.Cmd) {
	switch a.view {
	case viewDetail:
		return a.openSessionDetail()
	case viewSkillDetail:
		return a.openSkillDetail()
	case viewSkillsList:
		a.rescanSkills()
	}
	return a, nil
}

func (a App) handleResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	a.width = msg.Width
	a.height = msg.Height
	a.list.width = msg.Width
	a.list.height = msg.Height
	a.list.rebuildTable()
	a.skillsList.width = msg.Width
	a.skillsList.height = msg.Height
	a.skillsList.rebuildTable()
	a.permsList.width = msg.Width
	a.permsList.height = msg.Height
	a.permsList.rebuildTable()
	if a.view == viewDetail {
		a.detail.viewport.SetWidth(msg.Width)
		a.detail.viewport.SetHeight(msg.Height - 14)
	}
	if a.view == viewSkillDetail {
		a.skillDetail.viewport.SetWidth(msg.Width)
		a.skillDetail.viewport.SetHeight(msg.Height - 10)
	}
	return a, nil
}

// --- Tab cycling ---

var tabViewMap = map[Tab]viewState{
	TabSessions:    viewList,
	TabSkills:      viewSkillsList,
	TabPermissions: viewPermsList,
}

func (a App) nextTab() (tea.Model, tea.Cmd) {
	next := Tab((int(a.tab.active) + 1) % len(tabNames))
	a.tab.active = next
	a.view = tabViewMap[next]
	return a, nil
}

func (a App) prevTab() (tea.Model, tea.Cmd) {
	prev := int(a.tab.active) - 1
	if prev < 0 {
		prev = len(tabNames) - 1
	}
	a.tab.active = Tab(prev)
	a.view = tabViewMap[Tab(prev)]
	return a, nil
}

// --- Delete confirm (shared across all tabs) ---

func (a App) updateDeleteConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "y" || msg.String() == "Y" {
			a.executeDelete()
		}
		a.confirmDelete = false
		return a, nil
	}
	return a, nil
}

func (a *App) executeDelete() {
	switch a.view {
	case viewList, viewDetail:
		if s := a.list.selectedSession(); s != nil {
			claude.DeleteSession(a.claudeDir, *s)
			a.list.removeSession(s.ID)
			if a.view == viewDetail {
				a.view = viewList
			}
		}
	case viewSkillsList, viewSkillDetail:
		if sk := a.skillsList.selectedSkill(); sk != nil && !sk.ReadOnly {
			skillspkg.DeleteSkill(*sk)
			a.skillsList.removeSkill(sk.Name)
			if a.view == viewSkillDetail {
				a.view = viewSkillsList
			}
		}
	}
}

// --- Shared helpers ---

type editorFinishedMsg struct{}

type statusClearMsg struct{}

func statusClearCmd() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return statusClearMsg{}
	})
}

func openEditor(path string) tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	c := exec.Command(editor, path)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{}
	})
}

func replaceLastLine(content, replacement string) string {
	lines := strings.Split(content, "\n")
	if len(lines) > 0 {
		lines[len(lines)-1] = replacement
	}
	return strings.Join(lines, "\n")
}

func (a *App) rescanSkills() {
	skills, _ := skillspkg.ScanSkills(a.claudeDir)
	a.skillsList.skills = skills
	a.skillsList.applyFilter()
}

// --- View ---

func (a App) View() tea.View {
	var content string

	// Tab bar for list views
	switch a.view {
	case viewList, viewSkillsList, viewPermsList:
		content = a.tab.view(a.width)
	}

	switch a.view {
	case viewList:
		content += a.list.view()
	case viewDetail:
		content = a.detail.view()
	case viewSettings:
		content = a.settings.view()
	case viewSkillsList:
		content += a.skillsList.view()
	case viewSkillDetail:
		content = a.skillDetail.view()
	case viewPermsList:
		content += a.permsList.view()
	case viewPermsDetail:
		if a.addRule.active {
			content = a.addRule.view()
		} else {
			content = a.permsDetail.view()
		}
	}

	// Inline prompts (replace the help bar to avoid overflowing the terminal)
	if a.newSession {
		content = replaceLastLine(content, helpKeyStyle.Render("New session path: ")+a.newSessionDir.View())
	}
	if a.newSkill {
		content = replaceLastLine(content, helpKeyStyle.Render("New skill name: ")+a.newSkillInput.View())
	}
	if a.renaming {
		content = replaceLastLine(content, helpKeyStyle.Render("Rename: ")+a.renameInput.View())
	}

	// Export format picker (replaces help bar)
	if a.exportPicker {
		content = replaceLastLine(content,
			helpKeyStyle.Render("Export as: ")+
				helpKeyStyle.Render("m")+helpDescStyle.Render("arkdown")+
				helpSepStyle.Render(" • ")+
				helpKeyStyle.Render("h")+helpDescStyle.Render("tml")+
				helpSepStyle.Render(" • ")+
				helpKeyStyle.Render("esc")+helpDescStyle.Render(" cancel"))
	}

	// Status message (replaces help bar)
	if a.statusMsg != "" {
		content = replaceLastLine(content, valueStyle.Render(a.statusMsg))
	}

	// Delete confirmation (replaces help bar)
	if a.confirmDelete {
		var name string
		switch a.view {
		case viewList, viewDetail:
			if s := a.list.selectedSession(); s != nil {
				name = s.Name
			}
		case viewSkillsList, viewSkillDetail:
			if sk := a.skillsList.selectedSkill(); sk != nil {
				name = sk.Name
			}
		}
		content = replaceLastLine(content, deletePromptStyle.Render(
			fmt.Sprintf("Delete \"%s\"? This cannot be undone. (y/N)", name)))
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}
