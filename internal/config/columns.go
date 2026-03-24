package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// UserColumn represents a user-defined table column loaded from column.json.
type UserColumn struct {
	Name        string `json:"-"`
	Dir         string `json:"-"`
	Description string `json:"description"`
	Header      string `json:"header"`            // column header text
	Shell       string `json:"shell"`             // script filename to execute
	Width       int    `json:"width,omitempty"`   // column width (default 10)
	Timeout     int    `json:"timeout,omitempty"` // timeout in seconds (default 5)
}

func columnsDir() string { return filepath.Join(configDir, "columns") }

func ScanUserColumns() []UserColumn { return ScanUserColumnsFrom(columnsDir()) }

func ScanUserColumnsFrom(dir string) []UserColumn {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var cols []UserColumn
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if !ValidateCommandName(name) { // reuse same name validation
			continue
		}
		colDir := filepath.Join(dir, name)
		jsonPath := filepath.Join(colDir, "column.json")
		data, err := os.ReadFile(jsonPath)
		if err != nil {
			continue
		}
		var col UserColumn
		if err := json.Unmarshal(data, &col); err != nil {
			continue
		}
		col.Name = name
		col.Dir = colDir

		if col.Header == "" || col.Shell == "" {
			continue
		}
		shellPath := filepath.Join(colDir, col.Shell)
		if _, err := os.Stat(shellPath); err != nil {
			continue
		}
		if col.Width <= 0 {
			col.Width = 10
		}
		if col.Timeout <= 0 {
			col.Timeout = 5
		}

		cols = append(cols, col)
	}
	return cols
}

func ScaffoldColumn(name string) error {
	return ScaffoldColumnIn(columnsDir(), name)
}

func ScaffoldColumnIn(dir, name string) error {
	colDir := filepath.Join(dir, name)
	if err := os.MkdirAll(colDir, 0755); err != nil {
		return err
	}

	colJSON := map[string]any{
		"description": "TODO: describe what this column shows",
		"header":      name,
		"shell":       "run.sh",
		"width":       10,
		"timeout":     5,
	}
	data, _ := json.MarshalIndent(colJSON, "", "  ")
	if err := os.WriteFile(filepath.Join(colDir, "column.json"), data, 0644); err != nil {
		return err
	}

	script := fmt.Sprintf("#!/bin/bash\n# Receives session directory as $1\n# Output one line: the cell value\necho \"—\"\n")
	if err := os.WriteFile(filepath.Join(colDir, "run.sh"), []byte(script), 0755); err != nil {
		return err
	}

	return nil
}

func DeleteColumn(name string) error {
	return os.RemoveAll(filepath.Join(columnsDir(), name))
}

func UserColumnNames() []string {
	cols := ScanUserColumns()
	var names []string
	for _, c := range cols {
		names = append(names, c.Name)
	}
	return names
}
