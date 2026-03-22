package claude

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanSessions(t *testing.T) {
	tmpDir := t.TempDir()
	projDir := filepath.Join(tmpDir, "projects", "-Users-or-projects-myapp")
	if err := os.MkdirAll(projDir, 0o755); err != nil {
		t.Fatal(err)
	}

	jsonl := `{"type":"user","message":{"role":"user","content":"Add auth middleware"},"uuid":"u1","timestamp":"2026-03-22T14:30:00Z","cwd":"/Users/or/projects/myapp","gitBranch":"main"}
{"type":"assistant","message":{"role":"assistant","model":"claude-sonnet-4-6","content":[{"type":"text","text":"I will add the middleware..."}],"usage":{"input_tokens":5000,"output_tokens":1200,"cache_read_input_tokens":3000}},"uuid":"a1","timestamp":"2026-03-22T14:31:00Z"}
`
	if err := os.WriteFile(filepath.Join(projDir, "abc-123.jsonl"), []byte(jsonl), 0o644); err != nil {
		t.Fatal(err)
	}

	sessions, err := ScanSessions(tmpDir)
	if err != nil {
		t.Fatalf("ScanSessions returned error: %v", err)
	}

	if len(sessions) != 1 {
		t.Fatalf("got %d sessions, want 1", len(sessions))
	}

	s := sessions[0]
	if s.ID != "abc-123" {
		t.Errorf("ID = %q, want %q", s.ID, "abc-123")
	}
	if s.Name != "Add auth middleware" {
		t.Errorf("Name = %q, want %q", s.Name, "Add auth middleware")
	}
	if s.Directory != "/Users/or/projects/myapp" {
		t.Errorf("Directory = %q, want %q", s.Directory, "/Users/or/projects/myapp")
	}
	if s.Branch != "main" {
		t.Errorf("Branch = %q, want %q", s.Branch, "main")
	}
	if s.UserMsgs != 1 {
		t.Errorf("UserMsgs = %d, want 1", s.UserMsgs)
	}
	if s.AssistantMsgs != 1 {
		t.Errorf("AssistantMsgs = %d, want 1", s.AssistantMsgs)
	}
	if s.InputTokens != 5000 {
		t.Errorf("InputTokens = %d, want 5000", s.InputTokens)
	}
	if s.CacheTokens != 3000 {
		t.Errorf("CacheTokens = %d, want 3000", s.CacheTokens)
	}
	if s.OutputTokens != 1200 {
		t.Errorf("OutputTokens = %d, want 1200", s.OutputTokens)
	}
	if s.ModelID != "claude-sonnet-4-6" {
		t.Errorf("ModelID = %q, want %q", s.ModelID, "claude-sonnet-4-6")
	}
	if s.ProjectPath != "-Users-or-projects-myapp" {
		t.Errorf("ProjectPath = %q, want %q", s.ProjectPath, "-Users-or-projects-myapp")
	}
}

func TestScanSessionsEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	projDir := filepath.Join(tmpDir, "projects")
	if err := os.MkdirAll(projDir, 0o755); err != nil {
		t.Fatal(err)
	}

	sessions, err := ScanSessions(tmpDir)
	if err != nil {
		t.Fatalf("ScanSessions returned error: %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("got %d sessions, want 0", len(sessions))
	}
}

func TestNameTruncation(t *testing.T) {
	tmpDir := t.TempDir()
	projDir := filepath.Join(tmpDir, "projects", "-Users-or-projects-myapp")
	if err := os.MkdirAll(projDir, 0o755); err != nil {
		t.Fatal(err)
	}

	longMsg := "This is a very long message that exceeds eighty characters and should be truncated with an ellipsis at the end of it"
	jsonl := `{"type":"user","message":{"role":"user","content":"` + longMsg + `"},"uuid":"u1","timestamp":"2026-03-22T14:30:00Z","cwd":"/tmp","gitBranch":"dev"}
{"type":"assistant","message":{"role":"assistant","model":"claude-sonnet-4-6","content":[{"type":"text","text":"OK"}],"usage":{"input_tokens":100,"output_tokens":10,"cache_read_input_tokens":0}},"uuid":"a1","timestamp":"2026-03-22T14:31:00Z"}
`
	if err := os.WriteFile(filepath.Join(projDir, "long-name.jsonl"), []byte(jsonl), 0o644); err != nil {
		t.Fatal(err)
	}

	sessions, err := ScanSessions(tmpDir)
	if err != nil {
		t.Fatalf("ScanSessions returned error: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("got %d sessions, want 1", len(sessions))
	}

	s := sessions[0]
	// 80 runes + "..."
	runes := []rune(s.Name)
	if len(runes) != 83 {
		t.Errorf("Name length = %d runes, want 83 (80 + 3 for ...)", len(runes))
	}
	if s.Name[len(s.Name)-3:] != "..." {
		t.Errorf("Name should end with '...', got suffix %q", s.Name[len(s.Name)-3:])
	}
}
