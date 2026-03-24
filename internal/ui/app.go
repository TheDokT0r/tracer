package ui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	commandMode          bool
	cmdInput             commandInput
	cmdRegistry          *registry
	cmdHistory           []string
	pendingRescan        bool
	pendingCommandDelete string
	statusMsg            string
	width         int
	height        int
}

func NewApp(claudeDir string, sessions []model.Session, pins map[string]bool, cfg config.Config, renames map[string]string, skills []skillspkg.Skill, settingsFiles []ccsettings.SettingsFile) App {
	return App{
		claudeDir:   claudeDir,
		cfg:         cfg,
		renames:     renames,
		tab:         tabBar{active: TabSessions},
		list:        newListView(sessions, pins, cfg, 80, 24),
		skillsList:  newSkillsListView(skills, 80, 24),
		permsList:   newPermsListView(settingsFiles, 80, 24),
		view:        viewList,
		cmdRegistry: func() *registry {
			reg := defaultRegistry()
			loadUserCommands(reg)
			return reg
		}(),
		cmdHistory:  config.LoadHistory(),
	}
}

func (a App) Init() tea.Cmd {
	var cmds []tea.Cmd

	// Run autostart commands
	for _, uc := range config.ScanUserCommands() {
		if uc.Autostart && uc.Shell != "" && uc.Mode == "status" {
			uc := uc
			cmds = append(cmds, func() tea.Msg {
				return autostartMsg{cmd: uc}
			})
		}
	}

	// Populate custom columns async
	if len(a.list.columns) > 0 {
		cmds = append(cmds, func() tea.Msg {
			return columnTickMsg{}
		})
	}

	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}
	return nil
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case editorFinishedMsg:
		return a.handleEditorFinished()
	case userCommandFinishedMsg:
		a.rescanCommands()
		a.rescanColumns()
		return a, nil
	case columnResultMsg:
		m := msg.(columnResultMsg)
		if a.list.columnData[m.column] == nil {
			a.list.columnData[m.column] = make(map[string]string)
		}
		a.list.columnData[m.column][m.sessionID] = m.value
		a.list.rebuildTable()
		return a, nil
	case columnTickMsg:
		return a, a.populateColumns()
	case autostartMsg:
		result := runShellCommand(&a, msg.(autostartMsg).cmd, nil)
		if result != nil {
			return a, result
		}
		return a, nil
	case statusClearMsg:
		a.statusMsg = ""
		return a, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return a.handleResize(msg)
	}

	if a.commandMode {
		return a.updateCommand(msg)
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
	if a.pendingRescan {
		a.pendingRescan = false
		a.rescanCommands()
		a.rescanColumns()
		return a, nil
	}
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
	if a.pendingCommandDelete != "" {
		if strings.HasPrefix(a.pendingCommandDelete, "col:") {
			name := a.pendingCommandDelete[4:]
			config.DeleteColumn(name)
			a.statusMsg = "Deleted column: " + name
			a.pendingCommandDelete = ""
			a.rescanColumns()
			return
		}
		config.DeleteCommand(a.pendingCommandDelete)
		a.statusMsg = "Deleted command: " + a.pendingCommandDelete
		a.pendingCommandDelete = ""
		a.rescanCommands()
		return
	}
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

// --- Command mode ---

func (a *App) anyModalActive() bool {
	return a.confirmDelete || a.exportPicker || a.renaming ||
		a.newSession || a.newSkill || a.addRule.active ||
		a.list.filtering || a.skillsList.filtering || a.permsList.filtering
}

func (a App) enterCommandMode() (tea.Model, tea.Cmd) {
	a.commandMode = true
	a.cmdInput = newCommandInput(&a, a.cmdRegistry, a.cfg, a.view, a.cmdHistory)
	return a, nil
}

func (a App) updateCommand(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		execute, cancel, value := a.cmdInput.update(msg)
		if cancel {
			a.commandMode = false
			return a, nil
		}
		if execute {
			a.commandMode = false
			return a.executeCommand(value)
		}
	}
	return a, nil
}

func (a App) executeCommand(input string) (tea.Model, tea.Cmd) {
	input = strings.TrimSpace(input)
	if input == "" {
		return a, nil
	}

	cmd, args := a.cmdRegistry.resolve(input)
	if cmd == nil {
		a.statusMsg = "Unknown command: " + input
		return a, statusClearCmd()
	}

	if !cmd.availableIn(a.view) {
		a.statusMsg = "Command not available here: " + cmd.Name
		return a, statusClearCmd()
	}

	result := cmd.Run(&a, args)

	a.cmdHistory = config.AppendHistory(a.cmdHistory, input)
	config.SaveHistory(a.cmdHistory)

	if result != nil {
		return a, result
	}
	if a.statusMsg != "" {
		return a, statusClearCmd()
	}
	return a, nil
}

// --- Shared helpers ---

type editorFinishedMsg struct{}

type userCommandFinishedMsg struct{}

type autostartMsg struct {
	cmd config.UserCommand
}

type columnResultMsg struct {
	column    string
	sessionID string
	value     string
}

type columnTickMsg struct{}

type statusClearMsg struct{}

func isErrorStatus(msg string) bool {
	for _, prefix := range []string{
		"Unknown ", "Error:", "Usage:", "Invalid ", "Failed:",
		"Command not ", "Column not ", "Cannot ", "Claude not ",
		"No messages", "Export failed", "Alias loop", "not available",
	} {
		if strings.HasPrefix(msg, prefix) {
			return true
		}
	}
	return false
}

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

func (a *App) populateColumns() tea.Cmd {
	columns := config.ScanUserColumns()
	var cmds []tea.Cmd
	for _, col := range columns {
		if a.cfg.IsColumnHidden(col.Name) {
			continue
		}
		col := col
		for _, s := range a.list.sessions {
			s := s
			// Skip if already populated
			if data, ok := a.list.columnData[col.Name]; ok {
				if _, ok := data[s.ID]; ok {
					continue
				}
			}
			cmds = append(cmds, func() tea.Msg {
				scriptPath := filepath.Join(col.Dir, col.Shell)
				dir := s.Directory
				if dir == "" {
					dir = os.Getenv("HOME")
				}
				ctx, cancel := context.WithTimeout(context.Background(), time.Duration(col.Timeout)*time.Second)
				defer cancel()
				c := exec.CommandContext(ctx, scriptPath, dir)
				c.Dir = col.Dir
				c.Env = append(os.Environ(), "TRACER_COL_DIR="+col.Dir, "SESSION_DIR="+s.Directory, "SESSION_ID="+s.ID)
				out, err := c.Output()
				val := "—"
				if err == nil {
					line := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)[0]
					if line != "" {
						val = line
					}
				}
				return columnResultMsg{column: col.Name, sessionID: s.ID, value: val}
			})
		}
	}
	if len(cmds) > 0 {
		return tea.Batch(cmds...)
	}
	return nil
}

func (a *App) rescanColumns() {
	a.list.columns = config.ScanUserColumns()
	a.list.columnData = make(map[string]map[string]string)
	a.list.rebuildTable()
}

func (a *App) rescanCommands() {
	reg := defaultRegistry()
	loadUserCommands(reg)
	a.cmdRegistry = reg
	if a.commandMode {
		a.cmdInput.registry = reg
	}
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

	// Command palette (replaces help bar + dropdown above)
	if a.commandMode {
		dropdown := a.cmdInput.viewDropdown(a.width)
		inputLine := a.cmdInput.viewInput()
		if dropdown != "" {
			lines := strings.Split(content, "\n")
			dropLines := strings.Split(dropdown, "\n")
			needed := len(dropLines) + 1
			if len(lines) >= needed {
				lines = lines[:len(lines)-needed]
			}
			lines = append(lines, dropLines...)
			lines = append(lines, inputLine)
			content = strings.Join(lines, "\n")
		} else {
			content = replaceLastLine(content, inputLine)
		}
	}

	// Status message (shown on the line above the help bar)
	if a.statusMsg != "" {
		lines := strings.Split(content, "\n")
		if len(lines) >= 2 {
			style := valueStyle
			if isErrorStatus(a.statusMsg) {
				style = deletePromptStyle
			}
			lines[len(lines)-2] = style.Render(a.statusMsg)
			content = strings.Join(lines, "\n")
		}
	}

	// Delete confirmation (replaces help bar)
	if a.confirmDelete {
		var name string
		if a.pendingCommandDelete != "" {
			name = "command: " + a.pendingCommandDelete
		} else {
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
		}
		content = replaceLastLine(content, deletePromptStyle.Render(
			fmt.Sprintf("Delete \"%s\"? This cannot be undone. (y/N)", name)))
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}
