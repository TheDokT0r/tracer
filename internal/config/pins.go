package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

var pinsPath string

func init() {
	home, _ := os.UserHomeDir()
	pinsPath = filepath.Join(home, ".config", "tracer", "pins.json")
}

// LoadPins returns the set of pinned session IDs.
func LoadPins() map[string]bool {
	pins := make(map[string]bool)
	data, err := os.ReadFile(pinsPath)
	if err != nil {
		return pins
	}
	var ids []string
	if err := json.Unmarshal(data, &ids); err != nil {
		return pins
	}
	for _, id := range ids {
		pins[id] = true
	}
	return pins
}

// SavePins writes the pinned session IDs to disk.
func SavePins(pins map[string]bool) error {
	ids := make([]string, 0, len(pins))
	for id := range pins {
		ids = append(ids, id)
	}
	data, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(pinsPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(pinsPath, data, 0644)
}

// TogglePin adds or removes a session ID from pins. Returns new pinned state.
func TogglePin(pins map[string]bool, id string) bool {
	if pins[id] {
		delete(pins, id)
		return false
	}
	pins[id] = true
	return true
}
