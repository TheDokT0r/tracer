package ui

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	tea "charm.land/bubbletea/v2"
	"tracer/internal/claude"
	"tracer/internal/model"
)

type viewState int

const (
	viewList viewState = iota
	viewDetail
)

type App struct {
	claudeDir     string
	view          viewState
	list          listView
	detail        detailView
	confirmDelete bool
	width         int
	height        int
}

func NewApp(claudeDir string, sessions []model.Session) App {
	return App{
		claudeDir: claudeDir,
		list:      newListView(sessions, 80, 24),
		view:      viewList,
	}
}

func (a App) Init() tea.Cmd {
	return nil
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.list.width = msg.Width
		a.list.height = msg.Height
		a.list.rebuildTable()
		if a.view == viewDetail {
			a.detail.width = msg.Width
			a.detail.height = msg.Height
			a.detail.viewport.SetWidth(msg.Width)
			a.detail.viewport.SetHeight(msg.Height - 14)
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
	}
	return a, nil
}

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

		switch msg.String() {
		case "q", "ctrl+c":
			return a, tea.Quit
		case "/":
			a.list.filtering = true
			a.list.filter.Focus()
			return a, nil
		case "enter":
			return a.resumeSession()
		case "v":
			return a.openDetail()
		case "c":
			return a.copySessionID()
		case "d":
			if s := a.list.selectedSession(); s != nil {
				a.confirmDelete = true
			}
			return a, nil
		}
	}

	var cmd tea.Cmd
	a.list.table, cmd = a.list.table.Update(msg)
	return a, cmd
}

func (a App) updateDetail(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc", "q":
			a.view = viewList
			return a, nil
		case "enter":
			return a.resumeSession()
		case "c":
			return a.copySessionID()
		case "d":
			a.confirmDelete = true
			return a, nil
		case "ctrl+c":
			return a, tea.Quit
		}
	}

	var cmd tea.Cmd
	a.detail.viewport, cmd = a.detail.viewport.Update(msg)
	return a, cmd
}

func (a App) updateDeleteConfirm(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "y", "Y":
			s := a.list.selectedSession()
			if s != nil {
				claude.DeleteSession(a.claudeDir, *s)
				a.list.removeSession(s.ID)
				if a.view == viewDetail {
					a.view = viewList
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
		// Try xclip, fall back to xsel
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

func (a App) View() tea.View {
	var content string

	switch a.view {
	case viewList:
		content = a.list.view()
	case viewDetail:
		content = a.detail.view()
	}

	if a.confirmDelete {
		s := a.list.selectedSession()
		name := ""
		if s != nil {
			name = s.Name
		}
		content += "\n" + deletePromptStyle.Render(
			fmt.Sprintf("Delete \"%s\"? This cannot be undone. (y/N)", name))
	}

	v := tea.NewView(content)
	v.AltScreen = true
	return v
}
