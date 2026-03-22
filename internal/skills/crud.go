package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CreateSkill creates a new skill at ~/.claude/skills/{name}/SKILL.md with a template.
// Returns the path to the created file.
func CreateSkill(claudeDir, name, description string) (string, error) {
	// Sanitize name to kebab-case
	name = strings.ToLower(strings.ReplaceAll(name, " ", "-"))

	dir := filepath.Join(claudeDir, "skills", name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create skill directory: %w", err)
	}

	path := filepath.Join(dir, "SKILL.md")
	if _, err := os.Stat(path); err == nil {
		return "", fmt.Errorf("skill %q already exists", name)
	}

	// Title case the name for the heading
	title := strings.ReplaceAll(name, "-", " ")
	title = toTitleCase(title)

	content := fmt.Sprintf(`---
name: %s
description: %s
---

# %s

## Overview

[Describe what this skill does]
`, name, description, title)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write skill file: %w", err)
	}

	return path, nil
}

// toTitleCase capitalises the first letter of each word.
func toTitleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

// DeleteSkill removes a skill from disk. Only allows deleting user and command skills.
func DeleteSkill(skill Skill) error {
	if skill.ReadOnly {
		return fmt.Errorf("cannot delete read-only skill %q (source: %s)", skill.Name, skill.Source)
	}

	switch skill.Source {
	case SourceUser:
		// Delete the entire skill directory
		return os.RemoveAll(skill.Dir)
	case SourceCommand, SourceProject:
		// Delete the single .md file
		return os.Remove(skill.Path)
	default:
		return fmt.Errorf("cannot delete skill from source %q", skill.Source)
	}
}
