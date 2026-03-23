package config

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadHistory_Empty(t *testing.T) {
	dir := t.TempDir()
	h := loadHistoryFrom(filepath.Join(dir, "history"))
	if len(h) != 0 {
		t.Errorf("expected empty, got %d entries", len(h))
	}
}

func TestSaveAndLoadHistory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history")
	entries := []string{"delete", "pin", "sort date"}
	saveHistoryTo(path, entries)

	loaded := loadHistoryFrom(path)
	if len(loaded) != 3 {
		t.Fatalf("expected 3, got %d", len(loaded))
	}
	for i, e := range entries {
		if loaded[i] != e {
			t.Errorf("entry %d: expected %q, got %q", i, e, loaded[i])
		}
	}
}

func TestAppendHistory_Dedup(t *testing.T) {
	entries := []string{"pin", "delete"}
	entries = appendHistory(entries, "delete")
	if len(entries) != 2 {
		t.Errorf("expected 2 (deduped), got %d", len(entries))
	}
	entries = appendHistory(entries, "pin")
	if len(entries) != 3 {
		t.Errorf("expected 3, got %d", len(entries))
	}
}

func TestAppendHistory_MaxCap(t *testing.T) {
	var entries []string
	for i := 0; i < 150; i++ {
		entries = appendHistory(entries, strings.Repeat("x", i+1))
	}
	if len(entries) != maxHistoryEntries {
		t.Errorf("expected %d, got %d", maxHistoryEntries, len(entries))
	}
}
