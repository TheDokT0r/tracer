package ui

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"tracer/internal/claude"
	"tracer/internal/codex"
	"tracer/internal/config"
	"tracer/internal/gemini"
	"tracer/internal/model"
)

func (a App) updateSessionList(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			return a.nextTab()
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
			return a.openSessionDetail()
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
		case ":":
			if !a.anyModalActive() {
				return a.enterCommandMode()
			}
		}
	}

	var cmd tea.Cmd
	a.list.table, cmd = a.list.table.Update(msg)
	return a, cmd
}

func (a App) updateSessionDetail(msg tea.Msg) (tea.Model, tea.Cmd) {
	if a.renaming {
		return a.updateRename(msg)
	}
	if a.exportPicker {
		return a.updateExportPicker(msg)
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
		case "x":
			a.exportPicker = true
			return a, nil
		case "d":
			if a.cfg.ConfirmDelete {
				a.confirmDelete = true
			} else {
				if s := a.list.selectedSession(); s != nil {
					claude.DeleteSession(a.claudeDir, *s)
					a.list.removeSession(s.ID)
					a.view = viewList
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
	a.detail.viewport, cmd = a.detail.viewport.Update(msg)
	return a, cmd
}

// --- Session actions ---

func (a App) openSessionDetail() (tea.Model, tea.Cmd) {
	s := a.list.selectedSession()
	if s == nil {
		return a, nil
	}
	var messages []model.Message
	var err error
	switch s.Agent {
	case model.AgentCodex:
		messages, err = codex.LoadSessionDetail(s.FilePath, s)
	case model.AgentGemini:
		messages, err = gemini.LoadSessionDetail(s.FilePath, s)
	default:
		messages, err = claude.LoadSessionDetail(a.claudeDir, s)
	}
	if err != nil {
		return a, nil
	}
	a.detail = newDetailView(*s, messages, a.width, a.height)
	a.view = viewDetail
	return a, nil
}

func (a App) resumeSession() (tea.Model, tea.Cmd) {
	s := a.list.selectedSession()
	if s == nil {
		return a, nil
	}

	var bin string
	var args []string
	var err error

	switch s.Agent {
	case model.AgentCodex:
		bin, err = exec.LookPath("codex")
		if err != nil {
			return a, nil
		}
		args = []string{"resume", s.ID}
	case model.AgentGemini:
		a.statusMsg = "Gemini CLI does not support session resume"
		return a, statusClearCmd()
	default:
		bin, err = exec.LookPath("claude")
		if err != nil {
			return a, nil
		}
		args = []string{"--resume", s.ID}
		if a.cfg.Model != "" {
			args = append(args, "--model", a.cfg.Model)
		}
		if s.Name != "" {
			args = append(args, "--name", s.Name)
		}
	}

	c := exec.Command(bin, args...)
	c.Dir = s.Directory
	return a, tea.ExecProcess(c, func(err error) tea.Msg { return tea.Quit() })
}

func (a App) forkSession() (tea.Model, tea.Cmd) {
	s := a.list.selectedSession()
	if s == nil {
		return a, nil
	}

	var bin string
	var args []string
	var err error

	switch s.Agent {
	case model.AgentCodex:
		bin, err = exec.LookPath("codex")
		if err != nil {
			return a, nil
		}
		args = []string{"fork", s.ID}
	case model.AgentGemini:
		a.statusMsg = "Gemini CLI does not support session fork"
		return a, statusClearCmd()
	default:
		bin, err = exec.LookPath("claude")
		if err != nil {
			return a, nil
		}
		args = []string{"--resume", s.ID, "--fork-session"}
		if a.cfg.Model != "" {
			args = append(args, "--model", a.cfg.Model)
		}
		if s.Name != "" {
			args = append(args, "--name", s.Name)
		}
	}

	c := exec.Command(bin, args...)
	c.Dir = s.Directory
	return a, tea.ExecProcess(c, func(err error) tea.Msg { return tea.Quit() })
}

func (a App) copySessionID() (tea.Model, tea.Cmd) {
	s := a.list.selectedSession()
	if s == nil {
		return a, nil
	}
	copyToClipboard(s.ID)
	return a, nil
}

func (a App) updateExportPicker(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "m":
			a.exportPicker = false
			return a.exportMarkdown()
		case "h":
			a.exportPicker = false
			return a.exportHTML()
		case "esc":
			a.exportPicker = false
			return a, nil
		}
	}
	return a, nil
}

func (a App) exportMarkdown() (tea.Model, tea.Cmd) {
	if len(a.detail.messages) == 0 {
		a.statusMsg = "No messages to export"
		return a, statusClearCmd()
	}
	path, err := claude.ExportMarkdown(a.detail.session, a.detail.messages)
	if err != nil {
		a.statusMsg = "Export failed: " + err.Error()
		return a, statusClearCmd()
	}
	copyToClipboard(path)
	a.statusMsg = "Exported to " + path + " (copied)"
	return a, statusClearCmd()
}

func (a App) exportHTML() (tea.Model, tea.Cmd) {
	var messages []model.RichMessage

	// Rich conversation (with tool use, images, thinking) only available for Claude
	if a.detail.session.Agent == model.AgentClaude {
		messages, _ = claude.LoadRichConversation(a.claudeDir, a.detail.session)
	}

	// Fall back to simple messages for non-Claude or on error
	if len(messages) == 0 {
		for _, m := range a.detail.messages {
			if m.Content != "" {
				messages = append(messages, model.RichMessage{
					Role:      m.Role,
					Blocks:    []model.ContentBlock{{Type: "text", Text: m.Content}},
					Timestamp: m.Timestamp,
				})
			}
		}
	}

	if len(messages) == 0 {
		a.statusMsg = "No messages to export"
		return a, statusClearCmd()
	}
	path, err := claude.ExportHTML(a.detail.session, messages)
	if err != nil {
		a.statusMsg = "Export failed: " + err.Error()
		return a, statusClearCmd()
	}
	copyToClipboard(path)
	a.statusMsg = "Exported to " + path + " (copied)"
	return a, statusClearCmd()
}

func copyToClipboard(text string) {
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
	clipCmd.Stdin = strings.NewReader(text)
	clipCmd.Run()
}

func (a App) editSessionFile() (tea.Model, tea.Cmd) {
	s := a.list.selectedSession()
	if s == nil {
		return a, nil
	}
	path := filepath.Join(a.claudeDir, "projects", s.ProjectPath, s.ID+".jsonl")
	return a, openEditor(path)
}

func (a App) startNewSession() (tea.Model, tea.Cmd) {
	agents := a.getEnabledAgents()
	if len(agents) == 0 {
		a.statusMsg = "No agents enabled"
		return a, statusClearCmd()
	}
	a.enabledAgents = agents
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

func (a App) getEnabledAgents() []model.Agent {
	var agents []model.Agent
	if a.cfg.AgentClaude {
		agents = append(agents, model.AgentClaude)
	}
	if a.cfg.AgentCodex {
		agents = append(agents, model.AgentCodex)
	}
	if a.cfg.AgentGemini {
		agents = append(agents, model.AgentGemini)
	}
	return agents
}

func (a App) updateNewSession(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	// Agent picker mode
	if a.agentPicker {
		switch msg.String() {
		case "esc":
			a.agentPicker = false
			a.newSession = false
			return a, nil
		case "left", "h":
			if a.newSessionAgent > 0 {
				a.newSessionAgent--
			}
			return a, nil
		case "right", "l":
			if a.newSessionAgent < len(a.enabledAgents)-1 {
				a.newSessionAgent++
			}
			return a, nil
		case "enter":
			a.agentPicker = false
			a.newSession = false
			dir := strings.TrimSpace(a.newSessionDir.Value())
			if dir == "" {
				dir, _ = os.Getwd()
			}
			return a.launchNewSession(a.enabledAgents[a.newSessionAgent], dir)
		}
		return a, nil
	}

	// Dir input mode
	switch msg.String() {
	case "esc":
		a.newSession = false
		return a, nil
	case "enter":
		// If only one agent enabled, skip picker
		if len(a.enabledAgents) == 1 {
			dir := strings.TrimSpace(a.newSessionDir.Value())
			if dir == "" {
				dir, _ = os.Getwd()
			}
			a.newSession = false
			return a.launchNewSession(a.enabledAgents[0], dir)
		}
		// Show agent picker
		a.agentPicker = true
		a.newSessionAgent = 0
		return a, nil
	}
	var cmd tea.Cmd
	a.newSessionDir, cmd = a.newSessionDir.Update(msg)
	return a, cmd
}

func (a App) launchNewSession(agent model.Agent, dir string) (tea.Model, tea.Cmd) {
	var bin string
	var args []string
	var err error

	switch agent {
	case model.AgentCodex:
		bin, err = exec.LookPath("codex")
		if err != nil {
			a.statusMsg = "codex not found in PATH"
			return a, statusClearCmd()
		}
	case model.AgentGemini:
		bin, err = exec.LookPath("gemini")
		if err != nil {
			a.statusMsg = "gemini not found in PATH"
			return a, statusClearCmd()
		}
	default:
		bin, err = exec.LookPath("claude")
		if err != nil {
			a.statusMsg = "claude not found in PATH"
			return a, statusClearCmd()
		}
		if a.cfg.Model != "" {
			args = append(args, "--model", a.cfg.Model)
		}
	}

	c := exec.Command(bin, args...)
	c.Dir = dir
	return a, tea.ExecProcess(c, func(err error) tea.Msg { return tea.Quit() })
}

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
			if s := a.list.selectedSession(); s != nil {
				name := strings.TrimSpace(a.renameInput.Value())
				if name != "" {
					a.renames[s.ID] = name
					config.SaveRenames(a.renames)
					claude.WriteRename(a.claudeDir, *s, name)
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
