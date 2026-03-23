package claude

import (
	"encoding/json"
	"strings"
	"time"

	"tracer/internal/model"
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
	Type    string `json:"type"`
	Text    string `json:"text"`
	Thinking string `json:"thinking"`
}

// RichContentBlock captures all content block types including images and tool use.
type RichContentBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text"`
	Thinking  string          `json:"thinking"`
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Input     json.RawMessage `json:"input"`
	Content   json.RawMessage `json:"content"` // for tool_result
	ToolUseID string          `json:"tool_use_id"`
	Source    *ImageSource    `json:"source"`
}

type ImageSource struct {
	Type      string `json:"type"`       // "base64"
	MediaType string `json:"media_type"` // "image/png", etc.
	Data      string `json:"data"`       // base64 encoded
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

// RichContentBlocks parses the message content into typed blocks,
// extracting images from tool_result nested content.
func (e Entry) RichContentBlocks() []model.ContentBlock {
	if e.Message.Content == nil {
		return nil
	}

	// Try string content first
	var s string
	if err := json.Unmarshal(e.Message.Content, &s); err == nil {
		if s != "" {
			return []model.ContentBlock{{Type: "text", Text: s}}
		}
		return nil
	}

	// Parse as array of blocks
	var blocks []RichContentBlock
	if err := json.Unmarshal(e.Message.Content, &blocks); err != nil {
		return nil
	}

	var result []model.ContentBlock
	for _, b := range blocks {
		switch b.Type {
		case "text":
			if b.Text != "" {
				result = append(result, model.ContentBlock{Type: "text", Text: b.Text})
			}
		case "thinking":
			if b.Thinking != "" {
				result = append(result, model.ContentBlock{Type: "thinking", Text: b.Thinking})
			}
		case "image":
			if b.Source != nil && b.Source.Data != "" {
				result = append(result, model.ContentBlock{
					Type:      "image",
					MediaType: b.Source.MediaType,
					Data:      b.Source.Data,
				})
			}
		case "tool_use":
			inputStr := ""
			if b.Input != nil {
				inputStr = string(b.Input)
			}
			result = append(result, model.ContentBlock{
				Type:      "tool_use",
				ToolName:  b.Name,
				ToolInput: inputStr,
			})
		case "tool_result":
			// Tool results can contain nested content with images
			text := extractToolResultText(b.Content)
			if text != "" {
				result = append(result, model.ContentBlock{Type: "tool_result", Text: text, ToolName: b.ToolUseID})
			}
			// Extract nested images
			extractToolResultImages(b.Content, &result)
		}
	}
	return result
}

func extractToolResultText(raw json.RawMessage) string {
	if raw == nil {
		return ""
	}
	// Try as string
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	// Try as array of blocks
	var blocks []RichContentBlock
	if err := json.Unmarshal(raw, &blocks); err == nil {
		for _, b := range blocks {
			if b.Type == "text" && b.Text != "" {
				return b.Text
			}
		}
	}
	return ""
}

func extractToolResultImages(raw json.RawMessage, result *[]model.ContentBlock) {
	if raw == nil {
		return
	}
	var blocks []RichContentBlock
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return
	}
	for _, b := range blocks {
		if b.Type == "image" && b.Source != nil && b.Source.Data != "" {
			*result = append(*result, model.ContentBlock{
				Type:      "image",
				MediaType: b.Source.MediaType,
				Data:      b.Source.Data,
			})
		}
	}
}

func parseLine(data []byte) (Entry, error) {
	var e Entry
	err := json.Unmarshal(data, &e)
	return e, err
}
