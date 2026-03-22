package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
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
	renames       map[string]string
	renaming      bool
	renameInput   textinput.Model
	newSession    bool
	newSessionDir textinput.Model
	newSkill      bool
	newSkillInput textinput.Model
	confirmDelete bool
	width         int
	height        int
}

func NewApp(claudeDir string, sessions []model.Session, pins map[string]bool, cfg config.Config, renames map[string]string, skills []skillspkg.Skill) App {
	return App{
		claudeDir:  claudeDir,
		cfg:        cfg,
		renames:    renames,
		tab:        tabBar{active: TabSessions},
		list:       newListView(sessions, pins, cfg, 80, 24),
		skillsList: newSkillsListView(skills, 80, 24),
		view:       viewList,
	}
}

func (a App) Init() tea.Cmd {
	return nil
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case editorFinishedMsg:
		if a.view == viewDetail {
			return a.openDetail()
		}
		if a.view == viewSkillDetail {
			return a.openSkillDetail()
		}
		// Rescan skills after editing
		if a.view == viewSkillsList {
			a.rescanSkills()
		}
		return a, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.list.width = msg.Width
		a.list.height = msg.Height
		a.list.rebuildTable()
		a.skillsList.width = msg.Width
		a.skillsList.height = msg.Height
		a.skillsList.rebuildTable()
		if a.view == viewDetail {
			a.detail.width = msg.Width
			a.detail.height = msg.Height
			a.detail.viewport.SetWidth(msg.Width)
			a.detail.viewport.SetHeight(msg.Height - 14)
		}
		if a.view == viewSkillDetail {
			a.skillDetail.width = msg.Width
			a.skillDetail.height = msg.Height
			a.skillDetail.viewport.SetWidth(msg.Width)
			a.skillDetail.viewport.SetHeight(msg.Height - 10)
		}
		return a, nil
	}

	if a.confirmDelete {
		return a.updateDeleteConfirm(msg)
	}

	switch a.view {
	case viewList:
		return a.updateList(msg)
	case viewDetail:
		return a.updateDetail(msg)
	case viewSettings:
		return a.updateSettings(msg)
	case viewSkillsList:
		return a.updateSkillsList(msg)
	case viewSkillDetail:
		return a.updateSkillDetail(msg)
	}
	return a, nil
}

// --- Sessions list ---

func (a App) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if a.list.filtering {
			switch msg.String() {
			case "esc":
				a.list.filtering = false
				a.list.filter.SetValue("")
				a.list.applyFilter()
				return a, nil
			case "enter":
				a.list.filtering = false
				return a, nil
			default:
				var cmd tea.Cmd
				a.list.filter, cmd = a.list.filter.Update(msg)
				a.list.applyFilter()
				return a, cmd
			}
		}

		if a.newSession {
			return a.updateNewSession(msg)
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit
		case "tab":
			a.tab.active = TabSkills
			a.view = viewSkillsList
			return a, nil
		case "/":
			a.list.filtering = true
			a.list.filter.Focus()
			return a, nil
		case "n":
			return a.startNewSession()
		case "enter":
			return a.resumeSession()
		case "f":
			return a.forkSession()
		case "v":
			return a.openDetail()
		case "c":
			return a.copySessionID()
		case "s":
			a.settings = newSettingsView(a.cfg, a.width, a.height)
			a.view = viewSettings
			return a, nil
		case "p":
			if s := a.list.selectedSession(); s != nil {
				config.TogglePin(a.list.pins, s.ID)
				config.SavePins(a.list.pins)
				a.list.sortSessions()
				a.list.rebuildTable()
			}
			return a, nil
		case "d":
			if s := a.list.selectedSession(); s != nil {
				if a.cfg.ConfirmDelete {
					a.confirmDelete = true
				} else {
					claude.DeleteSession(a.claudeDir, *s)
					a.list.removeSession(s.ID)
				}
			}
			return a, nil
		}
	}

	var cmd tea.Cmd
	a.list.table, cmd = a.list.table.Update(msg)
	return a, cmd
}

// --- Session detail ---

func (a App) updateDetail(msg tea.Msg) (tea.Model, tea.Cmd) {
	if a.renaming {
		return a.updateRename(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "q":
			a.view = viewList
			return a, nil
		case "enter":
			return a.resumeSession()
		case "f":
			return a.forkSession()
		case "c":
			return a.copySessionID()
		case "r":
			return a.startRename()
		case "e":
			return a.editSessionFile()
		case "d":
			if a.cfg.ConfirmDelete {
				a.confirmDelete = true
			} else {
				s := a.list.selectedSession()
				if s != nil {
					claude.DeleteSession(a.claudeDir, *s)
					a.list.removeSession(s.ID)
					a.view = viewList
				}
			}
			return a, nil
		case "ctrl+c":
			return a, tea.Quit
		}
	}

	var cmd tea.Cmd
	a.detail.viewport, cmd = a.detail.viewport.Update(msg)
	return a, cmd
}

// --- Skills list ---

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
			a.tab.active = TabSessions
			a.view = viewList
			return a, nil
		case "shift+tab":
			a.tab.active = TabSessions
			a.view = viewList
			return a, nil
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
			sk := a.skillsList.selectedSkill()
			if sk != nil && !sk.ReadOnly {
				if a.cfg.ConfirmDelete {
					a.confirmDelete = true
				} else {
					skillspkg.DeleteSkill(*sk)
					a.skillsList.removeSkill(sk.Name)
				}
			}
			return a, nil
		}
	}

	var cmd tea.Cmd
	a.skillsList.table, cmd = a.skillsList.table.Update(msg)
	return a, cmd
}

// --- Skill detail ---

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
			sk := a.currentSkill()
			if sk != nil && !sk.ReadOnly {
				if a.cfg.ConfirmDelete {
					a.confirmDelete = true
				} else {
					skillspkg.DeleteSkill(*sk)
					a.skillsList.removeSkill(sk.Name)
					a.view = viewSkillsList
				}
			}
			return a, nil
		case "ctrl+c":
			return a, tea.Quit
		}
	}

	var cmd tea.Cmd
	a.skillDetail.viewport, cmd = a.skillDetail.viewport.Update(msg)
	return a, cmd
}

// --- Settings ---

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

// --- Delete confirm ---

func (a App) updateDeleteConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "y", "Y":
			switch a.view {
			case viewList, viewDetail:
				s := a.list.selectedSession()
				if s != nil {
					claude.DeleteSession(a.claudeDir, *s)
					a.list.removeSession(s.ID)
					if a.view == viewDetail {
						a.view = viewList
					}
				}
			case viewSkillsList, viewSkillDetail:
				sk := a.skillsList.selectedSkill()
				if sk != nil && !sk.ReadOnly {
					skillspkg.DeleteSkill(*sk)
					a.skillsList.removeSkill(sk.Name)
					if a.view == viewSkillDetail {
						a.view = viewSkillsList
					}
				}
			}
			a.confirmDelete = false
			return a, nil
		default:
			a.confirmDelete = false
			return a, nil
		}
	}
	return a, nil
}

// --- Session actions ---

func (a App) startRename() (tea.Model, tea.Cmd) {
	s := a.list.selectedSession()
	if s == nil {
		return a, nil
	}
	ti := textinput.New()
	ti.Placeholder = "new name..."
	ti.SetValue(s.Name)
	ti.Focus()
	ti.CharLimit = 80
	a.renameInput = ti
	a.renaming = true
	return a, nil
}

func (a App) updateRename(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			a.renaming = false
			return a, nil
		case "enter":
			s := a.list.selectedSession()
			if s != nil {
				name := strings.TrimSpace(a.renameInput.Value())
				if name != "" {
					a.renames[s.ID] = name
					config.SaveRenames(a.renames)
					for i := range a.list.sessions {
						if a.list.sessions[i].ID == s.ID {
							a.list.sessions[i].Name = name
							break
						}
					}
					s.Name = name
					a.detail.session.Name = name
					a.list.rebuildTable()
				}
			}
			a.renaming = false
			return a, nil
		}
	}
	var cmd tea.Cmd
	a.renameInput, cmd = a.renameInput.Update(msg)
	return a, cmd
}

func (a App) startNewSession() (tea.Model, tea.Cmd) {
	cwd, _ := os.Getwd()
	ti := textinput.New()
	ti.Placeholder = cwd
	ti.SetValue(cwd)
	ti.Focus()
	ti.CharLimit = 256
	a.newSessionDir = ti
	a.newSession = true
	return a, nil
}

func (a App) updateNewSession(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		a.newSession = false
		return a, nil
	case "enter":
		dir := strings.TrimSpace(a.newSessionDir.Value())
		if dir == "" {
			dir, _ = os.Getwd()
		}
		a.newSession = false
		return a.launchNewSession(dir)
	}
	var cmd tea.Cmd
	a.newSessionDir, cmd = a.newSessionDir.Update(msg)
	return a, cmd
}

func (a App) launchNewSession(dir string) (tea.Model, tea.Cmd) {
	claudeBin, err := exec.LookPath("claude")
	if err != nil {
		return a, nil
	}
	c := exec.Command(claudeBin)
	c.Dir = dir
	return a, tea.ExecProcess(c, func(err error) tea.Msg {
		return tea.Quit()
	})
}

func (a App) openDetail() (tea.Model, tea.Cmd) {
	s := a.list.selectedSession()
	if s == nil {
		return a, nil
	}
	claude.LoadSessionDetails(a.claudeDir, s)
	messages, err := claude.LoadConversation(a.claudeDir, *s)
	if err != nil {
		return a, nil
	}
	a.detail = newDetailView(*s, messages, a.width, a.height)
	a.view = viewDetail
	return a, nil
}

func (a App) forkSession() (tea.Model, tea.Cmd) {
	s := a.list.selectedSession()
	if s == nil {
		return a, nil
	}
	claudeBin, err := exec.LookPath("claude")
	if err != nil {
		return a, nil
	}
	c := exec.Command(claudeBin, "--resume", s.ID, "--fork-session")
	c.Dir = s.Directory
	return a, tea.ExecProcess(c, func(err error) tea.Msg {
		return tea.Quit()
	})
}

func (a App) resumeSession() (tea.Model, tea.Cmd) {
	s := a.list.selectedSession()
	if s == nil {
		return a, nil
	}
	claudeBin, err := exec.LookPath("claude")
	if err != nil {
		return a, nil
	}
	c := exec.Command(claudeBin, "--resume", s.ID)
	c.Dir = s.Directory
	return a, tea.ExecProcess(c, func(err error) tea.Msg {
		return tea.Quit()
	})
}

func (a App) copySessionID() (tea.Model, tea.Cmd) {
	s := a.list.selectedSession()
	if s == nil {
		return a, nil
	}
	var clipCmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		clipCmd = exec.Command("pbcopy")
	default:
		if _, err := exec.LookPath("xclip"); err == nil {
			clipCmd = exec.Command("xclip", "-selection", "clipboard")
		} else {
			clipCmd = exec.Command("xsel", "--clipboard", "--input")
		}
	}
	clipCmd.Stdin = strings.NewReader(s.ID)
	clipCmd.Run()
	return a, nil
}

type editorFinishedMsg struct{}

func (a App) editSessionFile() (tea.Model, tea.Cmd) {
	s := a.list.selectedSession()
	if s == nil {
		return a, nil
	}
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	path := filepath.Join(a.claudeDir, "projects", s.ProjectPath, s.ID+".jsonl")
	c := exec.Command(editor, path)
	return a, tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{}
	})
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
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	c := exec.Command(editor, sk.Path)
	return a, tea.ExecProcess(c, func(err error) tea.Msg {
		return editorFinishedMsg{}
	})
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
		// Open in editor
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi"
		}
		c := exec.Command(editor, path)
		return a, tea.ExecProcess(c, func(err error) tea.Msg {
			return editorFinishedMsg{}
		})
	}
	var cmd tea.Cmd
	a.newSkillInput, cmd = a.newSkillInput.Update(msg)
	return a, cmd
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
	case viewList, viewSkillsList:
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
	}

	if a.newSession {
		content += "\n" + helpKeyStyle.Render("New session path: ") + a.newSessionDir.View()
	}

	if a.newSkill {
		content += "\n" + helpKeyStyle.Render("New skill name: ") + a.newSkillInput.View()
	}

	if a.renaming {
		content += "\n" + helpKeyStyle.Render("Rename: ") + a.renameInput.View()
	}

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
		// Replace the last line (help bar) with the delete prompt
		lines := strings.Split(content, "\n")
		if len(lines) > 0 {
			lines[len(lines)-1] = deletePromptStyle.Render(
				fmt.Sprintf("Delete \"%s\"? This cannot be undone. (y/N)", name))
			content = strings.Join(lines, "\n")
		}
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}
