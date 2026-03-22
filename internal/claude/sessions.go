package claude

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"tracer/internal/model"
)

const maxLineBuffer = 1024 * 1024 // 1 MB

// ScanSessions walks {claudeDir}/projects/ looking for *.jsonl files.
// Each JSONL file is a session. Returns sessions sorted by most recent first.
func ScanSessions(claudeDir string) ([]model.Session, error) {
	projectsDir := filepath.Join(claudeDir, "projects")

	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, err
	}

	var sessions []model.Session

	for _, projEntry := range entries {
		if !projEntry.IsDir() {
			continue
		}
		projDir := filepath.Join(projectsDir, projEntry.Name())
		files, err := os.ReadDir(projDir)
		if err != nil {
			continue
		}
		for _, f := range files {
			if f.IsDir() || !strings.HasSuffix(f.Name(), ".jsonl") {
				continue
			}
			path := filepath.Join(projDir, f.Name())
			sess, err := parseSessionFile(path)
			if err != nil {
				continue
			}
			sess.ProjectPath = projEntry.Name()
			sessions = append(sessions, sess)
		}
	}

	// Apply renamed session names from history.jsonl
	renames := loadRenames(claudeDir)
	for i := range sessions {
		if name, ok := renames[sessions[i].ID]; ok {
			sessions[i].Name = name
		}
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].StartedAt.After(sessions[j].StartedAt)
	})

	return sessions, nil
}

// loadRenames scans history.jsonl for /rename entries and returns a map
// of sessionID -> last renamed name.
func loadRenames(claudeDir string) map[string]string {
	renames := make(map[string]string)

	f, err := os.Open(filepath.Join(claudeDir, "history.jsonl"))
	if err != nil {
		return renames
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, maxLineBuffer), maxLineBuffer)

	for scanner.Scan() {
		line := scanner.Bytes()
		var entry struct {
			Display   string `json:"display"`
			SessionID string `json:"sessionId"`
		}
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}
		if strings.HasPrefix(entry.Display, "/rename ") {
			name := strings.TrimPrefix(entry.Display, "/rename ")
			name = strings.TrimSpace(name)
			if name != "" && entry.SessionID != "" {
				renames[entry.SessionID] = name
			}
		}
	}

	return renames
}

// parseSessionFile reads a JSONL file and extracts session metadata.
func parseSessionFile(path string) (model.Session, error) {
	f, err := os.Open(path)
	if err != nil {
		return model.Session{}, err
	}
	defer f.Close()

	sess := model.Session{
		ID: strings.TrimSuffix(filepath.Base(path), ".jsonl"),
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, maxLineBuffer), maxLineBuffer)

	firstUser := true
	var lastAssistantEntry Entry

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		e, err := parseLine(line)
		if err != nil || !e.IsMessage() {
			continue
		}

		switch e.Type {
		case "user":
			sess.UserMsgs++
			sess.MessageCount++
			if firstUser {
				firstUser = false
				content := e.MessageContent()
				sess.Name = truncateName(content, 80)
				sess.Directory = e.CWD
				if e.GitBranch != "" {
				sess.Branch = e.GitBranch
			} else {
				sess.Branch = "-"
			}
				sess.StartedAt = e.Timestamp
			}
		case "assistant":
			sess.AssistantMsgs++
			sess.MessageCount++
			sess.OutputTokens += e.Message.Usage.OutputTokens
			lastAssistantEntry = e
		}
	}

	if err := scanner.Err(); err != nil {
		return model.Session{}, err
	}

	// Take input/cache tokens and model from last assistant message.
	// Total context = input_tokens + cache_creation_input_tokens + cache_read_input_tokens
	if lastAssistantEntry.Type == "assistant" {
		u := lastAssistantEntry.Message.Usage
		sess.InputTokens = u.InputTokens + u.CacheCreate + u.CacheReadTokens
		sess.CacheTokens = u.CacheReadTokens
		if lastAssistantEntry.Message.Model != "" {
			sess.ModelID = lastAssistantEntry.Message.Model
		}
	}

	return sess, nil
}

// LoadConversation reads a full JSONL file and returns all user/assistant
// messages with non-empty content.
func LoadConversation(claudeDir string, session model.Session) ([]model.Message, error) {
	path := filepath.Join(claudeDir, "projects", session.ProjectPath, session.ID+".jsonl")

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, maxLineBuffer), maxLineBuffer)

	var messages []model.Message

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		e, err := parseLine(line)
		if err != nil || !e.IsMessage() {
			continue
		}

		content := e.MessageContent()
		if content == "" {
			continue
		}

		messages = append(messages, model.Message{
			Role:      e.Type,
			Content:   content,
			Timestamp: e.Timestamp,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

// truncateName trims content to maxLen chars and replaces newlines with spaces.
func truncateName(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	if len([]rune(s)) > maxLen {
		return string([]rune(s)[:maxLen]) + "..."
	}
	return s
}
