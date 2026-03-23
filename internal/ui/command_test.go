package ui

import (
	"testing"
)

func TestParseInput_Simple(t *testing.T) {
	name, args := parseCommandInput("delete")
	if name != "delete" {
		t.Errorf("expected 'delete', got %q", name)
	}
	if len(args) != 0 {
		t.Errorf("expected 0 args, got %d", len(args))
	}
}

func TestParseInput_WithArgs(t *testing.T) {
	name, args := parseCommandInput("export html")
	if name != "export" {
		t.Errorf("expected 'export', got %q", name)
	}
	if len(args) != 1 || args[0] != "html" {
		t.Errorf("expected args ['html'], got %v", args)
	}
}

func TestResolve_Simple(t *testing.T) {
	reg := newRegistry([]Command{
		{Name: "delete", Contexts: []viewState{viewList}},
		{Name: "pin", Contexts: []viewState{viewList}},
	})
	cmd, args := reg.resolve("delete")
	if cmd == nil || cmd.Name != "delete" {
		t.Errorf("expected delete command")
	}
	if len(args) != 0 {
		t.Errorf("expected 0 args, got %v", args)
	}
}

func TestResolve_MultiWord(t *testing.T) {
	reg := newRegistry([]Command{
		{Name: "new", Contexts: []viewState{viewList}},
		{Name: "new skill", Contexts: []viewState{viewSkillsList}},
	})
	cmd, args := reg.resolve("new skill myskill")
	if cmd == nil || cmd.Name != "new skill" {
		t.Errorf("expected 'new skill', got %v", cmd)
	}
	if len(args) != 1 || args[0] != "myskill" {
		t.Errorf("expected ['myskill'], got %v", args)
	}
}

func TestResolve_MultiWordFallback(t *testing.T) {
	reg := newRegistry([]Command{
		{Name: "new", Contexts: []viewState{viewList}},
		{Name: "new skill", Contexts: []viewState{viewSkillsList}},
	})
	cmd, args := reg.resolve("new /tmp/foo")
	if cmd == nil || cmd.Name != "new" {
		t.Errorf("expected 'new', got %v", cmd)
	}
	if len(args) != 1 || args[0] != "/tmp/foo" {
		t.Errorf("expected ['/tmp/foo'], got %v", args)
	}
}

func TestMatch_Prefix(t *testing.T) {
	reg := newRegistry([]Command{
		{Name: "delete", Contexts: []viewState{viewList}},
		{Name: "pin", Contexts: []viewState{viewList}},
	})
	matches := reg.match("de", viewList)
	if len(matches) != 1 || matches[0].Name != "delete" {
		t.Errorf("expected [delete], got %v", matches)
	}
}

func TestMatch_ContextFilter(t *testing.T) {
	reg := newRegistry([]Command{
		{Name: "pin", Contexts: []viewState{viewList}},
		{Name: "delete", Contexts: []viewState{viewList, viewSkillsList}},
	})
	matches := reg.match("", viewSkillsList)
	for _, m := range matches {
		if m.Name == "pin" {
			t.Error("pin should not match in skills list context")
		}
	}
}
