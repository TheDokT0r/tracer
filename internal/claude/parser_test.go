package claude

import (
	"testing"
	"time"
)

func TestParseUserMessageStringContent(t *testing.T) {
	input := []byte(`{"type":"user","message":{"role":"user","content":"Add auth middleware"},"uuid":"abc","timestamp":"2026-03-22T14:30:00Z","cwd":"/Users/or/projects/my-app","gitBranch":"main"}`)

	e, err := parseLine(input)
	if err != nil {
		t.Fatalf("parseLine returned error: %v", err)
	}

	if e.Type != "user" {
		t.Errorf("Type = %q, want %q", e.Type, "user")
	}
	if !e.IsMessage() {
		t.Error("IsMessage() = false, want true")
	}
	if got := e.MessageContent(); got != "Add auth middleware" {
		t.Errorf("MessageContent() = %q, want %q", got, "Add auth middleware")
	}
	if e.CWD != "/Users/or/projects/my-app" {
		t.Errorf("CWD = %q, want %q", e.CWD, "/Users/or/projects/my-app")
	}
	if e.GitBranch != "main" {
		t.Errorf("GitBranch = %q, want %q", e.GitBranch, "main")
	}
	wantTime := time.Date(2026, 3, 22, 14, 30, 0, 0, time.UTC)
	if !e.Timestamp.Equal(wantTime) {
		t.Errorf("Timestamp = %v, want %v", e.Timestamp, wantTime)
	}
}

func TestParseAssistantMessage(t *testing.T) {
	input := []byte(`{"type":"assistant","message":{"role":"assistant","model":"claude-sonnet-4-6","content":[{"type":"text","text":"I will add the middleware..."}],"usage":{"input_tokens":5000,"output_tokens":1200,"cache_read_input_tokens":3000}},"uuid":"def","timestamp":"2026-03-22T14:31:00Z"}`)

	e, err := parseLine(input)
	if err != nil {
		t.Fatalf("parseLine returned error: %v", err)
	}

	if e.Type != "assistant" {
		t.Errorf("Type = %q, want %q", e.Type, "assistant")
	}
	if !e.IsMessage() {
		t.Error("IsMessage() = false, want true")
	}
	if got := e.MessageContent(); got != "I will add the middleware..." {
		t.Errorf("MessageContent() = %q, want %q", got, "I will add the middleware...")
	}
	if e.Message.Model != "claude-sonnet-4-6" {
		t.Errorf("Model = %q, want %q", e.Message.Model, "claude-sonnet-4-6")
	}
	if e.Message.Usage.InputTokens != 5000 {
		t.Errorf("InputTokens = %d, want %d", e.Message.Usage.InputTokens, 5000)
	}
	if e.Message.Usage.OutputTokens != 1200 {
		t.Errorf("OutputTokens = %d, want %d", e.Message.Usage.OutputTokens, 1200)
	}
	if e.Message.Usage.CacheReadTokens != 3000 {
		t.Errorf("CacheReadTokens = %d, want %d", e.Message.Usage.CacheReadTokens, 3000)
	}
}

func TestParseUserMessageContentBlocks(t *testing.T) {
	input := []byte(`{"type":"user","message":{"role":"user","content":[{"type":"text","text":"Fix the login bug"}]},"uuid":"ghi","timestamp":"2026-03-22T15:00:00Z","cwd":"/tmp"}`)

	e, err := parseLine(input)
	if err != nil {
		t.Fatalf("parseLine returned error: %v", err)
	}

	if e.Type != "user" {
		t.Errorf("Type = %q, want %q", e.Type, "user")
	}
	if got := e.MessageContent(); got != "Fix the login bug" {
		t.Errorf("MessageContent() = %q, want %q", got, "Fix the login bug")
	}
}

func TestParseNonMessageType(t *testing.T) {
	input := []byte(`{"type":"file-history-snapshot","messageId":"xyz"}`)

	e, err := parseLine(input)
	if err != nil {
		t.Fatalf("parseLine returned error: %v", err)
	}

	if e.Type != "file-history-snapshot" {
		t.Errorf("Type = %q, want %q", e.Type, "file-history-snapshot")
	}
	if e.IsMessage() {
		t.Error("IsMessage() = true, want false")
	}
	if got := e.MessageContent(); got != "" {
		t.Errorf("MessageContent() = %q, want empty string", got)
	}
}
