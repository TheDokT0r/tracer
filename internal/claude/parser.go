package claude

import (
	"encoding/json"
	"strings"
	"time"
)

type Entry struct {
	Type      string    `json:"type"`
	Message   RawMsg    `json:"message"`
	UUID      string    `json:"uuid"`
	Timestamp time.Time `json:"timestamp"`
	CWD       string    `json:"cwd"`
	GitBranch string    `json:"gitBranch"`
	SessionID string    `json:"sessionId"`
	Version   string    `json:"version"`
	IsMeta    bool      `json:"isMeta"`
}

type RawMsg struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
	Model   string          `json:"model"`
	Usage   Usage           `json:"usage"`
}

type Usage struct {
	InputTokens     int `json:"input_tokens"`
	OutputTokens    int `json:"output_tokens"`
	CacheCreate     int `json:"cache_creation_input_tokens"`
	CacheReadTokens int `json:"cache_read_input_tokens"`
}

type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (e Entry) IsMessage() bool {
	return e.Type == "user" || e.Type == "assistant"
}

// IsRealUserMessage returns true if this is a genuine user message,
// not a system/meta/command message.
func (e Entry) IsRealUserMessage() bool {
	if e.Type != "user" || e.IsMeta {
		return false
	}
	content := e.MessageContent()
	if content == "" {
		return false
	}
	// Skip XML-tagged system messages
	if strings.HasPrefix(content, "<") {
		return false
	}
	// Skip slash commands
	if strings.HasPrefix(content, "/") {
		return false
	}
	return true
}

func (e Entry) MessageContent() string {
	if e.Message.Content == nil {
		return ""
	}
	var s string
	if err := json.Unmarshal(e.Message.Content, &s); err == nil {
		return s
	}
	var blocks []ContentBlock
	if err := json.Unmarshal(e.Message.Content, &blocks); err == nil {
		for _, b := range blocks {
			if b.Type == "text" && b.Text != "" {
				return b.Text
			}
		}
	}
	return ""
}

func parseLine(data []byte) (Entry, error) {
	var e Entry
	err := json.Unmarshal(data, &e)
	return e, err
}
