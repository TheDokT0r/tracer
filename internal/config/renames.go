package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

var renamesPath string

func init() {
	home, _ := os.UserHomeDir()
	renamesPath = filepath.Join(home, ".config", "tracer", "renames.json")
}

// LoadRenames returns a map of sessionID -> custom name.
func LoadRenames() map[string]string {
	renames := make(map[string]string)
	data, err := os.ReadFile(renamesPath)
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
	if err := os.MkdirAll(filepath.Dir(renamesPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(renamesPath, data, 0644)
}
