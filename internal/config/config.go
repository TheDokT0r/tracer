package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

var configDir string

func init() {
	home := os.Getenv("HOME")
	if home == "" {
		home, _ = os.UserHomeDir()
	}
	if home == "" {
		home = "/tmp"
	}
	configDir = filepath.Join(home, ".config", "tracer")
}

type Config struct {
	Theme             string `json:"theme"`
	SortBy            string `json:"sort_by"`         // "date", "name", "directory"
	ShowDate          bool   `json:"show_date"`
	ShowDirectory     bool   `json:"show_directory"`
	ShowBranch        bool   `json:"show_branch"`
	ShowModel         bool   `json:"show_model"`
	Model             string `json:"model,omitempty"`
	ConfirmDelete     bool   `json:"confirm_delete"`
	AutoUpdate        bool   `json:"auto_update"`
	CmdDropdown       bool   `json:"cmd_dropdown"`
	CmdGhost          bool   `json:"cmd_ghost"`
	CmdMaxSuggestions int      `json:"cmd_max_suggestions"`
	HiddenColumns     []string `json:"hidden_columns,omitempty"`
}

func (c Config) IsColumnHidden(name string) bool {
	for _, h := range c.HiddenColumns {
		if h == name {
			return true
		}
	}
	return false
}

func (c *Config) ToggleColumn(name string) {
	for i, h := range c.HiddenColumns {
		if h == name {
			c.HiddenColumns = append(c.HiddenColumns[:i], c.HiddenColumns[i+1:]...)
			return
		}
	}
	c.HiddenColumns = append(c.HiddenColumns, name)
}

func DefaultConfig() Config {
	return Config{
		Theme:             "default",
		SortBy:            "date",
		ShowDate:          true,
		ShowDirectory:     true,
		ShowBranch:        true,
		ConfirmDelete:     true,
		CmdDropdown:       true,
		CmdMaxSuggestions: 8,
	}
}

func configPath() string { return filepath.Join(configDir, "config.json") }

func LoadConfig() Config {
	c := DefaultConfig()
	data, err := os.ReadFile(configPath())
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
	if c.CmdMaxSuggestions < 3 || c.CmdMaxSuggestions > 12 {
		c.CmdMaxSuggestions = 8
	}
	return c
}

func SaveConfig(c Config) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(configPath(), data, 0644)
}
