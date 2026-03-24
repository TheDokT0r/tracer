package ui

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"tracer/internal/config"
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
	Options  func(a *App) []Completion
	Required bool
}

// Completion is a single autocomplete suggestion with an optional description.
type Completion struct {
	Value       string
	Description string
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
				if len(cmd.Name) >= bestLen {
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

func (r *registry) completions(a *App, input string, ctx viewState) []Completion {
	cmd, args := r.resolve(input)
	if cmd == nil {
		matches := r.match(input, ctx)
		var comps []Completion
		for _, m := range matches {
			comps = append(comps, Completion{Value: m.Name, Description: m.Description})
		}
		return comps
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
		var filtered []Completion
		for _, opt := range options {
			if strings.HasPrefix(strings.ToLower(opt.Value), partial) {
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

// --- User command loading ---

func loadUserCommands(reg *registry) {
	userCmds := config.ScanUserCommands()
	for _, uc := range userCmds {
		cmd := buildUserCommand(uc, reg)
		reg.commands = append(reg.commands, cmd)
	}
}

func buildUserCommand(uc config.UserCommand, reg *registry) Command {
	var args []CommandArg
	for _, a := range uc.Args {
		arg := CommandArg{Name: a.Name, Required: a.Required}
		if len(a.Completions) > 0 {
			comps := make([]Completion, len(a.Completions))
			for i, c := range a.Completions {
				comps[i] = Completion{Value: c.Value, Description: c.Description}
			}
			arg.Options = func(_ *App) []Completion { return comps }
		}
		args = append(args, arg)
	}

	cmd := Command{
		Name:        uc.Name,
		Description: uc.Description,
		Contexts:    allExceptSettings,
		Args:        args,
	}

	if uc.Alias != "" {
		alias := uc.Alias
		cmd.Run = func(a *App, args []string) tea.Cmd {
			return resolveAlias(a, alias, args, 0)
		}
	} else {
		uc := uc // capture for closure
		cmd.Run = func(a *App, args []string) tea.Cmd {
			return runShellCommand(a, uc, args)
		}
	}

	return cmd
}

func resolveAlias(a *App, alias string, extraArgs []string, depth int) tea.Cmd {
	if depth > 10 {
		a.statusMsg = "Alias loop detected: " + alias
		return statusClearCmd()
	}
	input := alias
	if len(extraArgs) > 0 {
		input += " " + strings.Join(extraArgs, " ")
	}
	cmd, args := a.cmdRegistry.resolve(input)
	if cmd == nil {
		a.statusMsg = "Unknown command in alias: " + alias
		return statusClearCmd()
	}
	// Check if resolved command is also a user alias — recurse with depth+1
	for _, uc := range config.ScanUserCommands() {
		if uc.Name == cmd.Name && uc.Alias != "" {
			return resolveAlias(a, uc.Alias, args, depth+1)
		}
	}
	return cmd.Run(a, args)
}

func runShellCommand(a *App, uc config.UserCommand, args []string) tea.Cmd {
	scriptPath := filepath.Join(uc.Dir, uc.Shell)

	workDir := os.Getenv("HOME")
	if s := a.list.selectedSession(); s != nil && s.Directory != "" {
		workDir = s.Directory
	}

	if uc.Mode == "status" {
		c := exec.Command(scriptPath, args...)
		c.Dir = workDir
		c.Env = append(os.Environ(), "TRACER_CMD_DIR="+uc.Dir)
		out, err := c.CombinedOutput()
		if err != nil {
			a.statusMsg = "Error: " + strings.TrimSpace(string(out))
		} else {
			line := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)[0]
			if a.width > 0 && len(line) > a.width-2 {
				line = line[:a.width-2]
			}
			a.statusMsg = line
		}
		return statusClearCmd()
	}

	// Exec mode
	c := exec.Command(scriptPath, args...)
	c.Dir = workDir
	c.Env = append(os.Environ(), "TRACER_CMD_DIR="+uc.Dir)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return userCommandFinishedMsg{}
	})
}
