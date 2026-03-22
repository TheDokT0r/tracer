package ccsettings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// settingsJSON is the on-disk format (only fields we care about).
type settingsJSON struct {
	Permissions Permissions `json:"permissions"`
}

// ScanSettings finds all settings.json files: global + all projects.
func ScanSettings(claudeDir string) []SettingsFile {
	var files []SettingsFile

	// Global settings
	globalPath := filepath.Join(claudeDir, "settings.json")
	if sf, err := loadSettingsFile(globalPath, ScopeGlobal, ""); err == nil {
		files = append(files, sf)
	}

	// Scan project directories for .claude/settings.json and .claude/settings.local.json
	projectsDir := filepath.Join(claudeDir, "projects")
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return files
	}

	seen := make(map[string]bool)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		projectPath := decodeProjectPath(e.Name())
		if seen[projectPath] {
			continue
		}
		seen[projectPath] = true

		// Project settings (.claude/settings.json)
		projSettings := filepath.Join(projectPath, ".claude", "settings.json")
		if sf, err := loadSettingsFile(projSettings, ScopeProject, projectPath); err == nil {
			files = append(files, sf)
		}

		// Local settings (.claude/settings.local.json)
		localSettings := filepath.Join(projectPath, ".claude", "settings.local.json")
		if sf, err := loadSettingsFile(localSettings, ScopeLocal, projectPath); err == nil {
			files = append(files, sf)
		}
	}

	return files
}

func loadSettingsFile(path string, scope Scope, projectPath string) (SettingsFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SettingsFile{}, err
	}

	var sj settingsJSON
	if err := json.Unmarshal(data, &sj); err != nil {
		return SettingsFile{}, err
	}

	return SettingsFile{
		Scope:       scope,
		Path:        path,
		ProjectPath: projectPath,
		Permissions: sj.Permissions,
	}, nil
}

// SavePermissions writes updated permissions back to the settings file.
// Preserves all other fields in the JSON.
func SavePermissions(sf SettingsFile) error {
	// Read existing file to preserve other fields
	existing := make(map[string]interface{})
	data, err := os.ReadFile(sf.Path)
	if err == nil {
		json.Unmarshal(data, &existing)
	}

	// Update permissions
	perms := make(map[string]interface{})
	if sf.Permissions.Allow != nil {
		perms["allow"] = sf.Permissions.Allow
	}
	if sf.Permissions.Deny != nil {
		perms["deny"] = sf.Permissions.Deny
	}
	existing["permissions"] = perms

	out, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	os.MkdirAll(filepath.Dir(sf.Path), 0755)
	return os.WriteFile(sf.Path, out, 0644)
}

// AddRule adds a permission rule to a settings file.
func AddRule(sf *SettingsFile, list, rule string) {
	switch list {
	case "allow":
		sf.Permissions.Allow = appendUnique(sf.Permissions.Allow, rule)
	case "deny":
		sf.Permissions.Deny = appendUnique(sf.Permissions.Deny, rule)
	}
}

// RemoveRule removes a permission rule from a settings file.
func RemoveRule(sf *SettingsFile, list, rule string) {
	switch list {
	case "allow":
		sf.Permissions.Allow = removeStr(sf.Permissions.Allow, rule)
	case "deny":
		sf.Permissions.Deny = removeStr(sf.Permissions.Deny, rule)
	}
}

func appendUnique(slice []string, s string) []string {
	for _, v := range slice {
		if v == s {
			return slice
		}
	}
	return append(slice, s)
}

func removeStr(slice []string, s string) []string {
	result := make([]string, 0, len(slice))
	for _, v := range slice {
		if v != s {
			result = append(result, v)
		}
	}
	return result
}

// decodeProjectPath converts encoded dir name back to a path.
// e.g. "-Users-or-projects-Shapes" -> "/Users/or/projects/Shapes"
func decodeProjectPath(encoded string) string {
	if len(encoded) == 0 {
		return ""
	}
	// First char "-" becomes "/", rest of "-" become "/"
	parts := strings.Split(encoded, "-")
	return "/" + strings.Join(parts[1:], "/")
}
