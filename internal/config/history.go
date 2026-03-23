package config

import (
	"os"
	"path/filepath"
	"strings"
)

const maxHistoryEntries = 100

func historyPath() string { return filepath.Join(configDir, "history") }

func LoadHistory() []string  { return loadHistoryFrom(historyPath()) }

func SaveHistory(entries []string) { saveHistoryTo(historyPath(), entries) }

func AppendHistory(entries []string, cmd string) []string {
	return appendHistory(entries, cmd)
}

func loadHistoryFrom(path string) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	var result []string
	for _, line := range lines {
		if line != "" {
			result = append(result, line)
		}
	}
	return result
}

func saveHistoryTo(path string, entries []string) {
	os.MkdirAll(filepath.Dir(path), 0755)
	os.WriteFile(path, []byte(strings.Join(entries, "\n")+"\n"), 0644)
}

func appendHistory(entries []string, cmd string) []string {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return entries
	}
	if len(entries) > 0 && entries[len(entries)-1] == cmd {
		return entries
	}
	entries = append(entries, cmd)
	if len(entries) > maxHistoryEntries {
		entries = entries[len(entries)-maxHistoryEntries:]
	}
	return entries
}
