package ui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"tracer/internal/config"
)

type commandInput struct {
	input          textinput.Model
	app            *App
	registry       *registry
	cfg            config.Config
	ctx            viewState
	suggestions    []Command
	argSuggestions []string
	selected       int
	history        []string
	historyIdx     int
	savedInput     string
	active         bool
}

func newCommandInput(app *App, reg *registry, cfg config.Config, ctx viewState, history []string) commandInput {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Focus()
	ti.CharLimit = 256
	ti.Prompt = ":"
	return commandInput{
		input:      ti,
		app:        app,
		registry:   reg,
		cfg:        cfg,
		ctx:        ctx,
		history:    history,
		historyIdx: -1,
		selected:   0,
		active:     true,
	}
}

func (ci *commandInput) update(msg tea.KeyPressMsg) (execute bool, cancel bool, value string) {
	k := msg.String()

	switch k {
	case "esc":
		return false, true, ""

	case "enter":
		input := ci.input.Value()
		if cmd, _ := ci.registry.resolve(input); cmd != nil {
			return true, false, input
		}
		if ci.selected >= 0 && ci.dropdownVisible() {
			ci.acceptSuggestion()
			return false, false, ""
		}
		return true, false, input

	case "tab":
		ci.acceptSuggestion()
		return false, false, ""

	case "up":
		if ci.dropdownVisible() && len(ci.allSuggestions()) > 0 {
			if ci.selected > 0 {
				ci.selected--
			}
		} else {
			ci.historyUp()
		}
		return false, false, ""

	case "down":
		if ci.dropdownVisible() && len(ci.allSuggestions()) > 0 {
			if ci.selected < len(ci.allSuggestions())-1 {
				ci.selected++
			}
		} else {
			ci.historyDown()
		}
		return false, false, ""

	case "backspace":
		if ci.input.Value() == "" {
			return false, true, ""
		}
		ci.input, _ = ci.input.Update(msg)
		ci.refreshSuggestions()
		return false, false, ""

	default:
		ci.input, _ = ci.input.Update(msg)
		ci.refreshSuggestions()
		ci.historyIdx = -1
		return false, false, ""
	}
}

func (ci *commandInput) refreshSuggestions() {
	input := ci.input.Value()
	ci.selected = 0

	cmd, _ := ci.registry.resolve(input)
	if cmd != nil && strings.Contains(input, " ") {
		ci.suggestions = nil
		ci.argSuggestions = ci.registry.completions(ci.app, input, ci.ctx)
	} else {
		ci.suggestions = ci.registry.match(input, ci.ctx)
		ci.argSuggestions = nil
	}
}

func (ci *commandInput) allSuggestions() []string {
	if len(ci.argSuggestions) > 0 {
		return ci.argSuggestions
	}
	var names []string
	for _, s := range ci.suggestions {
		names = append(names, s.Name)
	}
	return names
}

func (ci *commandInput) dropdownVisible() bool {
	return ci.cfg.CmdDropdown && len(ci.allSuggestions()) > 0
}

func (ci *commandInput) acceptSuggestion() {
	sugs := ci.allSuggestions()
	if len(sugs) == 0 {
		return
	}
	idx := ci.selected
	if idx < 0 || idx >= len(sugs) {
		idx = 0
	}
	suggestion := sugs[idx]

	if len(ci.argSuggestions) > 0 {
		cmd, _ := ci.registry.resolve(ci.input.Value())
		if cmd != nil {
			cmdWords := len(strings.Fields(cmd.Name))
			parts := strings.Fields(ci.input.Value())
			if len(parts) > cmdWords {
				parts = parts[:cmdWords]
			}
			parts = append(parts, suggestion)
			ci.input.SetValue(strings.Join(parts, " ") + " ")
		}
	} else {
		ci.input.SetValue(suggestion + " ")
	}
	ci.input.CursorEnd()
	ci.refreshSuggestions()
}

func (ci *commandInput) historyUp() {
	if len(ci.history) == 0 {
		return
	}
	if ci.historyIdx == -1 {
		ci.savedInput = ci.input.Value()
		ci.historyIdx = len(ci.history) - 1
	} else if ci.historyIdx > 0 {
		ci.historyIdx--
	}
	ci.input.SetValue(ci.history[ci.historyIdx])
	ci.input.CursorEnd()
}

func (ci *commandInput) historyDown() {
	if ci.historyIdx == -1 {
		return
	}
	if ci.historyIdx < len(ci.history)-1 {
		ci.historyIdx++
		ci.input.SetValue(ci.history[ci.historyIdx])
	} else {
		ci.historyIdx = -1
		ci.input.SetValue(ci.savedInput)
	}
	ci.input.CursorEnd()
}

func (ci *commandInput) ghostText() string {
	if !ci.cfg.CmdGhost {
		return ""
	}
	sugs := ci.allSuggestions()
	if len(sugs) == 0 {
		return ""
	}
	idx := ci.selected
	if idx < 0 || idx >= len(sugs) {
		idx = 0
	}
	suggestion := sugs[idx]
	current := ci.input.Value()
	if strings.HasPrefix(suggestion, current) {
		return suggestion[len(current):]
	}
	return ""
}

func (ci *commandInput) viewInput() string {
	ghost := ci.ghostText()
	line := ci.input.View()
	if ghost != "" {
		line += dimmedStyle.Render(ghost)
	}
	return line
}

func (ci *commandInput) viewDropdown(width int) string {
	if !ci.dropdownVisible() {
		return ""
	}

	allSugs := ci.allSuggestions()
	max := ci.cfg.CmdMaxSuggestions
	if max <= 0 {
		max = 8
	}

	// Calculate window
	start := 0
	if len(allSugs) > max {
		start = ci.selected - max/2
		if start < 0 {
			start = 0
		}
		if start+max > len(allSugs) {
			start = len(allSugs) - max
		}
	}
	end := start + max
	if end > len(allSugs) {
		end = len(allSugs)
	}
	visible := allSugs[start:end]

	selectedBg := lipgloss.NewStyle().
		Background(purple).
		Foreground(white).
		Bold(true)

	normalName := lipgloss.NewStyle().Foreground(white)

	var lines []string
	for i, sug := range visible {
		actualIdx := start + i
		name := sug
		desc := ""

		if len(ci.argSuggestions) == 0 {
			// It's a command name — find description from suggestions slice
			cmdIdx := actualIdx
			if cmdIdx < len(ci.suggestions) {
				desc = ci.suggestions[cmdIdx].Description
			}
		}

		nameStr := fmt.Sprintf("  %-20s", name)
		if desc != "" {
			nameStr += " " + desc
		}
		if len(nameStr) > width-2 && width > 4 {
			nameStr = nameStr[:width-2]
		}

		if actualIdx == ci.selected {
			lines = append(lines, selectedBg.Render(nameStr))
		} else {
			if desc != "" {
				namePart := fmt.Sprintf("  %-20s ", name)
				lines = append(lines, normalName.Render(namePart)+dimmedStyle.Render(desc))
			} else {
				lines = append(lines, normalName.Render(nameStr))
			}
		}
	}

	return strings.Join(lines, "\n")
}
