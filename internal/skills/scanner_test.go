package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanUserSkills(t *testing.T) {
	dir := t.TempDir()

	// Create skills/my-skill/SKILL.md with frontmatter
	skillDir := filepath.Join(dir, "skills", "my-skill")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\nname: My Skill\ndescription: A test skill\n---\nBody content here.\n"
	skillPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	skills, err := ScanSkills(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}

	sk := skills[0]
	if sk.Name != "My Skill" {
		t.Errorf("expected name 'My Skill', got %q", sk.Name)
	}
	if sk.Description != "A test skill" {
		t.Errorf("expected description 'A test skill', got %q", sk.Description)
	}
	if sk.Source != SourceUser {
		t.Errorf("expected source %q, got %q", SourceUser, sk.Source)
	}
	if sk.Path != skillPath {
		t.Errorf("expected path %q, got %q", skillPath, sk.Path)
	}
	if sk.ReadOnly {
		t.Error("user skill should not be read-only")
	}
}

func TestScanCommands(t *testing.T) {
	dir := t.TempDir()

	// Create commands/my-cmd.md with frontmatter
	cmdDir := filepath.Join(dir, "commands")
	if err := os.MkdirAll(cmdDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\nname: My Command\ndescription: Runs something\n---\nInstructions.\n"
	cmdPath := filepath.Join(cmdDir, "my-cmd.md")
	if err := os.WriteFile(cmdPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	skills, err := ScanSkills(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}

	sk := skills[0]
	if sk.Name != "My Command" {
		t.Errorf("expected name 'My Command', got %q", sk.Name)
	}
	if sk.Description != "Runs something" {
		t.Errorf("expected description 'Runs something', got %q", sk.Description)
	}
	if sk.Source != SourceCommand {
		t.Errorf("expected source %q, got %q", SourceCommand, sk.Source)
	}
	if sk.Path != cmdPath {
		t.Errorf("expected path %q, got %q", cmdPath, sk.Path)
	}
}

func TestParseFrontmatter(t *testing.T) {
	t.Run("valid frontmatter", func(t *testing.T) {
		data := []byte("---\nname: Hello\ndescription: World\n---\nBody\n")
		name, desc := parseFrontmatter(data)
		if name != "Hello" {
			t.Errorf("expected name 'Hello', got %q", name)
		}
		if desc != "World" {
			t.Errorf("expected description 'World', got %q", desc)
		}
	})

	t.Run("no frontmatter", func(t *testing.T) {
		data := []byte("Just plain content\n")
		name, desc := parseFrontmatter(data)
		if name != "" {
			t.Errorf("expected empty name, got %q", name)
		}
		if desc != "" {
			t.Errorf("expected empty description, got %q", desc)
		}
	})

	t.Run("empty description", func(t *testing.T) {
		data := []byte("---\nname: OnlyName\ndescription:\n---\nBody\n")
		name, desc := parseFrontmatter(data)
		if name != "OnlyName" {
			t.Errorf("expected name 'OnlyName', got %q", name)
		}
		if desc != "" {
			t.Errorf("expected empty description, got %q", desc)
		}
	})
}

func TestScanEmpty(t *testing.T) {
	dir := t.TempDir()

	skills, err := ScanSkills(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(skills) != 0 {
		t.Errorf("expected 0 skills, got %d", len(skills))
	}
}
