package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

var configPath string

func init() {
	home, _ := os.UserHomeDir()
	configPath = filepath.Join(home, ".config", "tracer", "config.json")
}

type Config struct {
	Theme         string `json:"theme"`
	SortBy        string `json:"sort_by"`         // "date", "name", "directory"
	ShowDate      bool   `json:"show_date"`
	ShowDirectory bool   `json:"show_directory"`
	ShowBranch    bool   `json:"show_branch"`
	ConfirmDelete bool   `json:"confirm_delete"`
	AutoUpdate    bool   `json:"auto_update"`
}

func DefaultConfig() Config {
	return Config{
		Theme:         "default",
		SortBy:        "date",
		ShowDate:      true,
		ShowDirectory: true,
		ShowBranch:    true,
		ConfirmDelete: true,
	}
}

func LoadConfig() Config {
	c := DefaultConfig()
	data, err := os.ReadFile(configPath)
	if err != nil {
		return c
	}
	json.Unmarshal(data, &c)
	if c.Theme == "" {
		c.Theme = "default"
	}
	if c.SortBy == "" {
		c.SortBy = "date"
	}
	return c
}

func SaveConfig(c Config) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}
