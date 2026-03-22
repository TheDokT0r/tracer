package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func renamesPath() string { return filepath.Join(configDir, "renames.json") }

// LoadRenames returns a map of sessionID -> custom name.
func LoadRenames() map[string]string {
	renames := make(map[string]string)
	data, err := os.ReadFile(renamesPath())
	if err != nil {
		return renames
	}
	json.Unmarshal(data, &renames)
	return renames
}

// SaveRenames writes the renames map to disk.
func SaveRenames(renames map[string]string) error {
	data, err := json.MarshalIndent(renames, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(renamesPath(), data, 0644)
}
