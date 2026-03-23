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
	"tracer/internal/config"
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
	claude.LoadSessionDetails(a.claudeDir, s)
	messages, err := claude.LoadConversation(a.claudeDir, *s)
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
	claudeBin, err := exec.LookPath("claude")
	if err != nil {
		return a, nil
	}
	c := exec.Command(claudeBin, "--resume", s.ID)
	c.Dir = s.Directory
	return a, tea.ExecProcess(c, func(err error) tea.Msg { return tea.Quit() })
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
	messages, err := claude.LoadRichConversation(a.claudeDir, a.detail.session)
	if err != nil || len(messages) == 0 {
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
		claudeBin, err := exec.LookPath("claude")
		if err != nil {
			return a, nil
		}
		c := exec.Command(claudeBin)
		c.Dir = dir
		return a, tea.ExecProcess(c, func(err error) tea.Msg { return tea.Quit() })
	}
	var cmd tea.Cmd
	a.newSessionDir, cmd = a.newSessionDir.Update(msg)
	return a, cmd
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
