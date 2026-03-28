package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// UserTheme is the on-disk format for a custom theme.
type UserTheme struct {
	Primary  string `json:"primary"`
	Accent   string `json:"accent"`
	Text     string `json:"text"`
	Bright   string `json:"bright"`
	Muted    string `json:"muted"`
	Dim      string `json:"dim"`
	Red      string `json:"red"`
	Green    string `json:"green"`
	SelectBg string `json:"select_bg"`
	SelectFg string `json:"select_fg"`
}

// ScanUserThemes reads all JSON files from ~/.config/tracer/themes/.
// Returns a map of theme name -> UserTheme.
func ScanUserThemes() map[string]UserTheme {
	themes := make(map[string]UserTheme)
	dir := filepath.Join(configDir, "themes")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return themes
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".json")
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		var t UserTheme
		if json.Unmarshal(data, &t) == nil && t.Primary != "" {
			themes[name] = t
		}
	}
	return themes
}

// ScaffoldTheme creates a theme JSON file with default colors.
func ScaffoldTheme(name string) (string, error) {
	dir := filepath.Join(configDir, "themes")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, name+".json")
	t := UserTheme{
		Primary:  "#7D56F4",
		Accent:   "#7D56F4",
		Text:     "#FAFAFA",
		Bright:   "#FFFFFF",
		Muted:    "#626262",
		Dim:      "#444444",
		Red:      "#FF4444",
		Green:    "#44FF44",
		SelectBg: "#7D56F4",
		SelectFg: "#FFFFFF",
	}
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return "", err
	}
	return path, os.WriteFile(path, data, 0644)
}

// ThemePath returns the path to a custom theme file.
func ThemePath(name string) string {
	return filepath.Join(configDir, "themes", name+".json")
}
