package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

var commandNameRe = regexp.MustCompile(`^[a-z0-9-]+$`)

// UserCommand represents a user-defined command loaded from command.json.
type UserCommand struct {
	Name        string           `json:"-"`
	Dir         string           `json:"-"`
	Description string           `json:"description"`
	Alias       string           `json:"alias,omitempty"`
	Shell       string           `json:"shell,omitempty"`
	Mode        string           `json:"mode,omitempty"` // "status" or "exec"
	Args        []UserCommandArg `json:"args,omitempty"`
	Autostart   bool             `json:"autostart,omitempty"`
}

type UserCommandArg struct {
	Name        string              `json:"name"`
	Required    bool                `json:"required"`
	Completions []UserCommandOption `json:"completions,omitempty"`
}

type UserCommandOption struct {
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
}

func ValidateCommandName(name string) bool {
	return name != "" && commandNameRe.MatchString(name)
}

func commandsDir() string { return filepath.Join(configDir, "commands") }

func ScanUserCommands() []UserCommand { return ScanUserCommandsFrom(commandsDir()) }

func ScanUserCommandsFrom(dir string) []UserCommand {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var cmds []UserCommand
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if !ValidateCommandName(name) {
			continue
		}
		cmdDir := filepath.Join(dir, name)
		jsonPath := filepath.Join(cmdDir, "command.json")
		data, err := os.ReadFile(jsonPath)
		if err != nil {
			continue
		}
		var cmd UserCommand
		if err := json.Unmarshal(data, &cmd); err != nil {
			continue
		}
		cmd.Name = name
		cmd.Dir = cmdDir

		if cmd.Description == "" {
			continue
		}
		hasAlias := cmd.Alias != ""
		hasShell := cmd.Shell != ""
		if hasAlias == hasShell {
			continue
		}
		if hasShell {
			shellPath := filepath.Join(cmdDir, cmd.Shell)
			if _, err := os.Stat(shellPath); err != nil {
				continue
			}
		}
		if cmd.Mode == "" && hasShell {
			cmd.Mode = "exec"
		}

		cmds = append(cmds, cmd)
	}
	return cmds
}

func ScaffoldCommand(name string) error {
	return ScaffoldCommandIn(commandsDir(), name)
}

func ScaffoldCommandIn(dir, name string) error {
	cmdDir := filepath.Join(dir, name)
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		return err
	}

	cmdJSON := map[string]any{
		"description": "TODO: describe what this command does",
		"shell":       "run.sh",
		"mode":        "exec",
	}
	data, _ := json.MarshalIndent(cmdJSON, "", "  ")
	if err := os.WriteFile(filepath.Join(cmdDir, "command.json"), data, 0644); err != nil {
		return err
	}

	script := fmt.Sprintf("#!/bin/bash\n# TODO: implement this command\necho \"Hello from %s\"\n", name)
	if err := os.WriteFile(filepath.Join(cmdDir, "run.sh"), []byte(script), 0755); err != nil {
		return err
	}

	return nil
}

func DeleteCommand(name string) error {
	return DeleteCommandIn(commandsDir(), name)
}

func DeleteCommandIn(dir, name string) error {
	return os.RemoveAll(filepath.Join(dir, name))
}

func UserCommandNames() []string {
	cmds := ScanUserCommands()
	var names []string
	for _, c := range cmds {
		names = append(names, c.Name)
	}
	return names
}
