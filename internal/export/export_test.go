package export

import (
	"os"
	"strings"
	"testing"
	"time"

	"tracer/internal/model"
)

func TestExportHTML_Structure(t *testing.T) {
	session := model.Session{
		ID:            "test-session-id",
		Name:          "Test Session",
		Directory:     "/tmp/test",
		StartedAt:     time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC),
		MessageCount:  2,
		UserMsgs:      1,
		AssistantMsgs: 1,
	}

	messages := []model.RichMessage{
		{
			Role: "user",
			Blocks: []model.ContentBlock{
				{Type: "text", Text: "Hello world"},
			},
			Timestamp: time.Date(2026, 3, 24, 10, 0, 0, 0, time.UTC),
		},
		{
			Role: "assistant",
			Blocks: []model.ContentBlock{
				{Type: "text", Text: "**Hi there!**\n\nHow can I help?"},
			},
			Timestamp: time.Date(2026, 3, 24, 10, 0, 1, 0, time.UTC),
		},
	}

	path, err := ExportHTML(session, messages)
	if err != nil {
		t.Fatalf("ExportHTML failed: %v", err)
	}
	defer os.Remove(path)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read export: %v", err)
	}
	content := string(data)

	// No sticky header
	if strings.Contains(content, `class="header"`) {
		t.Error("should not have sticky header div")
	}

	// Has inline chat header
	if !strings.Contains(content, `class="chat-header"`) {
		t.Error("should have inline chat-header")
	}

	// Both messages have bubbles
	if !strings.Contains(content, `class="bubble"`) {
		t.Error("messages should have bubble wrapper")
	}

	// User message on the right (flex-end via .msg.user)
	if !strings.Contains(content, `class="msg user"`) {
		t.Error("should have user message class")
	}

	// Assistant message on the left
	if !strings.Contains(content, `class="msg assistant"`) {
		t.Error("should have assistant message class")
	}

	// User text: escaped, not markdown-rendered
	if strings.Contains(content, `<p>Hello world</p>`) {
		t.Error("user text should not be markdown-rendered")
	}
	if !strings.Contains(content, `Hello world`) {
		t.Error("user text should be present")
	}

	// Assistant text: markdown-rendered
	if !strings.Contains(content, `<strong>Hi there!</strong>`) {
		t.Error("assistant text should be markdown-rendered with <strong>")
	}
}

func TestExportHTML_XSSPrevention(t *testing.T) {
	session := model.Session{
		ID:        "xss-test",
		Name:      "XSS Test",
		StartedAt: time.Now(),
	}

	messages := []model.RichMessage{
		{
			Role: "user",
			Blocks: []model.ContentBlock{
				{Type: "text", Text: "<script>alert('xss')</script>"},
			},
		},
		{
			Role: "assistant",
			Blocks: []model.ContentBlock{
				{Type: "text", Text: "Safe response"},
			},
		},
	}

	path, err := ExportHTML(session, messages)
	if err != nil {
		t.Fatalf("ExportHTML failed: %v", err)
	}
	defer os.Remove(path)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read export: %v", err)
	}
	content := string(data)

	if strings.Contains(content, "<script>alert") {
		t.Error("user script tags should be escaped")
	}
}

func TestExportHTML_ThinkingAndToolBlocks(t *testing.T) {
	session := model.Session{
		ID:        "blocks-test",
		Name:      "Blocks Test",
		StartedAt: time.Now(),
	}

	messages := []model.RichMessage{
		{
			Role: "assistant",
			Blocks: []model.ContentBlock{
				{Type: "thinking", Text: "let me think about this..."},
				{Type: "tool_use", ToolName: "Read", ToolInput: `{"path":"/tmp/file.txt"}`},
				{Type: "tool_result", Text: "file contents here"},
				{Type: "text", Text: "Here's what I found."},
			},
		},
	}

	path, err := ExportHTML(session, messages)
	if err != nil {
		t.Fatalf("ExportHTML failed: %v", err)
	}
	defer os.Remove(path)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read export: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, `class="detail-block thinking"`) {
		t.Error("should have thinking block")
	}
	if !strings.Contains(content, `class="detail-block tool-use"`) {
		t.Error("should have tool-use block")
	}
	if !strings.Contains(content, `class="detail-block tool-result"`) {
		t.Error("should have tool-result block")
	}
}
