package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
)

type Command struct {
	Name        string
	Args        []CommandArg
	Description string
	Contexts    []viewState
	Run         func(a *App, args []string) tea.Cmd
}

type CommandArg struct {
	Name     string
	Options  func(a *App) []string
	Required bool
}

type registry struct {
	commands []Command
}

func newRegistry(cmds []Command) *registry {
	return &registry{commands: cmds}
}

func (r *registry) resolve(input string) (*Command, []string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, nil
	}
	var best *Command
	var bestLen int
	for i := range r.commands {
		cmd := &r.commands[i]
		if strings.HasPrefix(input, cmd.Name) {
			rest := input[len(cmd.Name):]
			if rest == "" || rest[0] == ' ' {
				if len(cmd.Name) > bestLen {
					best = cmd
					bestLen = len(cmd.Name)
				}
			}
		}
	}
	if best == nil {
		return nil, nil
	}
	rest := strings.TrimSpace(input[bestLen:])
	var args []string
	if rest != "" {
		args = strings.Fields(rest)
	}
	return best, args
}

func (r *registry) match(prefix string, ctx viewState) []Command {
	prefix = strings.TrimSpace(strings.ToLower(prefix))
	var results []Command
	for _, cmd := range r.commands {
		if !cmd.availableIn(ctx) {
			continue
		}
		if prefix == "" || strings.HasPrefix(cmd.Name, prefix) {
			results = append(results, cmd)
		}
	}
	return results
}

func (r *registry) completions(a *App, input string, ctx viewState) []string {
	cmd, args := r.resolve(input)
	if cmd == nil {
		matches := r.match(input, ctx)
		var names []string
		for _, m := range matches {
			names = append(names, m.Name)
		}
		return names
	}
	var argIdx int
	if strings.HasSuffix(input, " ") || len(args) == 0 {
		argIdx = len(args)
	} else {
		argIdx = len(args) - 1
	}
	if argIdx >= len(cmd.Args) {
		return nil
	}
	arg := cmd.Args[argIdx]
	if arg.Options == nil {
		return nil
	}
	options := arg.Options(a)
	if argIdx < len(args) {
		partial := strings.ToLower(args[argIdx])
		var filtered []string
		for _, opt := range options {
			if strings.HasPrefix(strings.ToLower(opt), partial) {
				filtered = append(filtered, opt)
			}
		}
		return filtered
	}
	return options
}

func (c Command) availableIn(ctx viewState) bool {
	for _, v := range c.Contexts {
		if v == ctx {
			return true
		}
	}
	return false
}

func parseCommandInput(input string) (string, []string) {
	parts := strings.Fields(strings.TrimSpace(input))
	if len(parts) == 0 {
		return "", nil
	}
	return parts[0], parts[1:]
}
