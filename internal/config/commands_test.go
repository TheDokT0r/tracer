package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateName(t *testing.T) {
	valid := []string{"deploy", "gs", "my-cmd", "a1b2"}
	for _, n := range valid {
		if !ValidateCommandName(n) {
			t.Errorf("expected %q to be valid", n)
		}
	}
	invalid := []string{"", "My Cmd", "deploy!", "UPPER", "has space", "a_b", "a.b"}
	for _, n := range invalid {
		if ValidateCommandName(n) {
			t.Errorf("expected %q to be invalid", n)
		}
	}
}

func TestScanUserCommands_ValidAlias(t *testing.T) {
	dir := t.TempDir()
	cmdDir := filepath.Join(dir, "q")
	os.MkdirAll(cmdDir, 0755)
	data, _ := json.Marshal(map[string]any{"description": "Quit shortcut", "alias": "quit"})
	os.WriteFile(filepath.Join(cmdDir, "command.json"), data, 0644)

	cmds := ScanUserCommandsFrom(dir)
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	if cmds[0].Name != "q" || cmds[0].Alias != "quit" {
		t.Errorf("unexpected command: %+v", cmds[0])
	}
}

func TestScanUserCommands_ValidShell(t *testing.T) {
	dir := t.TempDir()
	cmdDir := filepath.Join(dir, "deploy")
	os.MkdirAll(cmdDir, 0755)
	data, _ := json.Marshal(map[string]any{"description": "Deploy app", "shell": "run.sh", "mode": "exec"})
	os.WriteFile(filepath.Join(cmdDir, "command.json"), data, 0644)
	os.WriteFile(filepath.Join(cmdDir, "run.sh"), []byte("#!/bin/bash\necho hi"), 0755)

	cmds := ScanUserCommandsFrom(dir)
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	if cmds[0].Shell != "run.sh" || cmds[0].Mode != "exec" {
		t.Errorf("unexpected: %+v", cmds[0])
	}
}

func TestScanUserCommands_MissingDescription(t *testing.T) {
	dir := t.TempDir()
	cmdDir := filepath.Join(dir, "bad")
	os.MkdirAll(cmdDir, 0755)
	data, _ := json.Marshal(map[string]any{"alias": "quit"})
	os.WriteFile(filepath.Join(cmdDir, "command.json"), data, 0644)

	cmds := ScanUserCommandsFrom(dir)
	if len(cmds) != 0 {
		t.Errorf("expected 0 (invalid), got %d", len(cmds))
	}
}

func TestScanUserCommands_BothAliasAndShell(t *testing.T) {
	dir := t.TempDir()
	cmdDir := filepath.Join(dir, "bad")
	os.MkdirAll(cmdDir, 0755)
	data, _ := json.Marshal(map[string]any{"description": "Bad", "alias": "quit", "shell": "run.sh"})
	os.WriteFile(filepath.Join(cmdDir, "command.json"), data, 0644)

	cmds := ScanUserCommandsFrom(dir)
	if len(cmds) != 0 {
		t.Errorf("expected 0 (invalid), got %d", len(cmds))
	}
}

func TestScanUserCommands_NeitherAliasNorShell(t *testing.T) {
	dir := t.TempDir()
	cmdDir := filepath.Join(dir, "bad")
	os.MkdirAll(cmdDir, 0755)
	data, _ := json.Marshal(map[string]any{"description": "Bad"})
	os.WriteFile(filepath.Join(cmdDir, "command.json"), data, 0644)

	cmds := ScanUserCommandsFrom(dir)
	if len(cmds) != 0 {
		t.Errorf("expected 0 (invalid), got %d", len(cmds))
	}
}

func TestScanUserCommands_InvalidName(t *testing.T) {
	dir := t.TempDir()
	cmdDir := filepath.Join(dir, "BAD NAME")
	os.MkdirAll(cmdDir, 0755)
	data, _ := json.Marshal(map[string]any{"description": "Bad", "alias": "quit"})
	os.WriteFile(filepath.Join(cmdDir, "command.json"), data, 0644)

	cmds := ScanUserCommandsFrom(dir)
	if len(cmds) != 0 {
		t.Errorf("expected 0 (invalid name), got %d", len(cmds))
	}
}

func TestScaffoldCommand(t *testing.T) {
	dir := t.TempDir()
	err := ScaffoldCommandIn(dir, "my-cmd")
	if err != nil {
		t.Fatalf("scaffold failed: %v", err)
	}
	cmdDir := filepath.Join(dir, "my-cmd")
	if _, err := os.Stat(filepath.Join(cmdDir, "command.json")); err != nil {
		t.Error("command.json not created")
	}
	if info, err := os.Stat(filepath.Join(cmdDir, "run.sh")); err != nil {
		t.Error("run.sh not created")
	} else if info.Mode()&0111 == 0 {
		t.Error("run.sh not executable")
	}
}

func TestDeleteCommand(t *testing.T) {
	dir := t.TempDir()
	ScaffoldCommandIn(dir, "doomed")
	if _, err := os.Stat(filepath.Join(dir, "doomed")); err != nil {
		t.Fatal("setup failed")
	}
	DeleteCommandIn(dir, "doomed")
	if _, err := os.Stat(filepath.Join(dir, "doomed")); !os.IsNotExist(err) {
		t.Error("directory should be deleted")
	}
}
