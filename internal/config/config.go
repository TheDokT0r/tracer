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
	Theme string `json:"theme"`
}

func LoadConfig() Config {
	c := Config{Theme: "default"}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return c
	}
	json.Unmarshal(data, &c)
	if c.Theme == "" {
		c.Theme = "default"
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
