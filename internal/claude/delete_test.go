package claude

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"tracer/internal/model"
)

func TestDeleteSession(t *testing.T) {
	tmpDir := t.TempDir()

	// Create session JSONL file
	projDir := filepath.Join(tmpDir, "projects", "-Users-or-projects-app")
	if err := os.MkdirAll(projDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projDir, "sess-1.jsonl"), []byte(`{"type":"user"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create session subdirectory with subagent file
	subDir := filepath.Join(projDir, "sess-1", "subagents")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "agent.jsonl"), []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create file-history
	fileHistDir := filepath.Join(tmpDir, "file-history", "sess-1")
	if err := os.MkdirAll(fileHistDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fileHistDir, "abc@v1"), []byte("old content"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create tasks
	tasksDir := filepath.Join(tmpDir, "tasks", "sess-1")
	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tasksDir, "1.json"), []byte(`{"task":1}`), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create history.jsonl: 2 entries for sess-1, 1 for sess-2
	historyLines := strings.Join([]string{
		`{"sessionId":"sess-1","type":"user","message":{"role":"user","content":"first"}}`,
		`{"sessionId":"sess-2","type":"user","message":{"role":"user","content":"other"}}`,
		`{"sessionId":"sess-1","type":"assistant","message":{"role":"assistant","content":"reply"}}`,
	}, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "history.jsonl"), []byte(historyLines), 0o644); err != nil {
		t.Fatal(err)
	}

	// Delete sess-1
	sess := model.Session{
		ID:          "sess-1",
		ProjectPath: "-Users-or-projects-app",
	}
	if err := DeleteSession(tmpDir, sess); err != nil {
		t.Fatalf("DeleteSession returned error: %v", err)
	}

	// Verify JSONL file is gone
	if _, err := os.Stat(filepath.Join(projDir, "sess-1.jsonl")); !os.IsNotExist(err) {
		t.Error("sess-1.jsonl should be deleted")
	}

	// Verify subdirectory is gone
	if _, err := os.Stat(filepath.Join(projDir, "sess-1")); !os.IsNotExist(err) {
		t.Error("sess-1/ subdirectory should be deleted")
	}

	// Verify file-history is gone
	if _, err := os.Stat(fileHistDir); !os.IsNotExist(err) {
		t.Error("file-history/sess-1/ should be deleted")
	}

	// Verify tasks dir is gone
	if _, err := os.Stat(tasksDir); !os.IsNotExist(err) {
		t.Error("tasks/sess-1/ should be deleted")
	}

	// Verify history.jsonl only has sess-2 entry
	data, err := os.ReadFile(filepath.Join(tmpDir, "history.jsonl"))
	if err != nil {
		t.Fatalf("reading history.jsonl: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 1 {
		t.Fatalf("history.jsonl has %d lines, want 1", len(lines))
	}
	if !strings.Contains(lines[0], `"sess-2"`) {
		t.Errorf("remaining history line should be sess-2, got: %s", lines[0])
	}
}

func TestDeleteSessionMissingFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Only create the projects dir so the path is valid but no session files exist
	projDir := filepath.Join(tmpDir, "projects", "-Users-or-projects-app")
	if err := os.MkdirAll(projDir, 0o755); err != nil {
		t.Fatal(err)
	}

	sess := model.Session{
		ID:          "nonexistent",
		ProjectPath: "-Users-or-projects-app",
	}

	// Should not error on missing files
	if err := DeleteSession(tmpDir, sess); err != nil {
		t.Fatalf("DeleteSession should not error for missing files: %v", err)
	}
}
