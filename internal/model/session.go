package model

import "time"

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
	InputTokens   int
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
	pct := float64(s.InputTokens) / float64(max)
	if pct > 1.0 {
		return 1.0
	}
	return pct
}
