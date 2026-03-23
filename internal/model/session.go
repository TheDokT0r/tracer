package model

import (
	"strings"
	"time"
)

// ContextWindows maps model ID prefixes to their max token counts.
var ContextWindows = map[string]int{
	"claude-opus-4":     200000,
	"claude-sonnet-4":   200000,
	"claude-haiku-4":    200000,
	"claude-opus-4-6":   200000,
	"claude-sonnet-4-6": 200000,
	"claude-haiku-4-5":  200000,
	"claude-3-5-sonnet": 200000,
	"claude-3-5-haiku":  200000,
}

const DefaultContextWindow = 200000
const ExtendedContextWindow = 1000000

// Session holds metadata for one Claude Code session.
type Session struct {
	ID            string
	Name          string    // First user message, truncated
	Directory     string    // Working directory
	Branch        string    // Git branch
	StartedAt     time.Time
	MessageCount  int       // Total messages
	UserMsgs      int
	AssistantMsgs int
	ContextTokens int // Total: input + cache_create + cache_read
	CacheTokens   int
	OutputTokens  int
	ModelID       string // For determining context window
	ProjectPath   string // Encoded project path (for file location)
}

// Message is a simplified conversation entry for the detail view.
type Message struct {
	Role      string // "user" or "assistant"
	Content   string
	Timestamp time.Time
}

// ContentBlock represents a single content block in a rich message.
type ContentBlock struct {
	Type      string // "text", "image", "tool_use", "tool_result", "thinking"
	Text      string // for text, thinking, tool_result blocks
	MediaType string // for image blocks (e.g. "image/png")
	Data      string // base64 data for image blocks
	ToolName  string // for tool_use blocks
	ToolInput string // for tool_use blocks (JSON string)
}

// RichMessage holds the full content blocks of a conversation entry.
type RichMessage struct {
	Role      string // "user" or "assistant"
	Blocks    []ContentBlock
	Timestamp time.Time
}

// Text returns the concatenated text content of all text blocks.
func (m RichMessage) Text() string {
	var parts []string
	for _, b := range m.Blocks {
		if b.Type == "text" && b.Text != "" {
			parts = append(parts, b.Text)
		}
	}
	return strings.Join(parts, "\n")
}

// HasImages returns true if the message contains any image blocks.
func (m RichMessage) HasImages() bool {
	for _, b := range m.Blocks {
		if b.Type == "image" {
			return true
		}
	}
	return false
}

// MaxContextTokens returns the context window size for this session's model.
func (s Session) MaxContextTokens() int {
	for prefix, tokens := range ContextWindows {
		if len(s.ModelID) >= len(prefix) && s.ModelID[:len(prefix)] == prefix {
			return tokens
		}
	}
	return DefaultContextWindow
}

// ContextPercent returns how full the context window is (0.0 to 1.0).
func (s Session) ContextPercent() float64 {
	max := s.MaxContextTokens()
	if max == 0 {
		return 0
	}
	pct := float64(s.ContextTokens) / float64(max)
	if pct > 1.0 {
		return 1.0
	}
	return pct
}
