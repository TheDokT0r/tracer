package claude

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
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

	// Collect all session file paths
	type sessionPath struct {
		path        string
		projectPath string
	}
	var paths []sessionPath

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
			paths = append(paths, sessionPath{
				path:        filepath.Join(projDir, f.Name()),
				projectPath: projEntry.Name(),
			})
		}
	}

	// Parse in parallel
	var mu sync.Mutex
	var wg sync.WaitGroup
	sessions := make([]model.Session, 0, len(paths))

	sem := make(chan struct{}, 16) // limit concurrency
	for _, sp := range paths {
		wg.Add(1)
		go func(sp sessionPath) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			sess, err := scanSessionHead(sp.path)
			if err != nil {
				return
			}
			sess.ProjectPath = sp.projectPath

			mu.Lock()
			sessions = append(sessions, sess)
			mu.Unlock()
		}(sp)
	}
	wg.Wait()

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

// scanSessionHead reads only the first user message from a JSONL file.
// This is fast — it stops reading as soon as it has the name, dir, branch, and date.
func scanSessionHead(path string) (model.Session, error) {
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

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		e, err := parseLine(line)
		if err != nil {
			continue
		}

		// Pick up cwd/branch/timestamp/model from first entry that has them
		if sess.Directory == "" && e.CWD != "" {
			sess.Directory = e.CWD
		}
		if sess.Branch == "" && e.GitBranch != "" && e.GitBranch != "HEAD" {
			sess.Branch = e.GitBranch
		}
		if sess.StartedAt.IsZero() && !e.Timestamp.IsZero() {
			sess.StartedAt = e.Timestamp
		}
		if sess.ModelID == "" && e.Type == "assistant" && e.Message.Model != "" {
			sess.ModelID = e.Message.Model
		}

		if !e.IsRealUserMessage() {
			continue
		}

		sess.Name = truncateName(e.MessageContent(), 80)
		if sess.Branch == "" {
			sess.Branch = "-"
		}
		return sess, nil
	}

	if err := scanner.Err(); err != nil {
		return model.Session{}, err
	}

	// No real user message found — skip this session
	if sess.Name == "" {
		return model.Session{}, fmt.Errorf("no user message found")
	}
	if sess.Branch == "" {
		sess.Branch = "-"
	}

	return sess, nil
}

// LoadSessionDetail reads the JSONL file in a single pass, populating both
// session metadata (token counts, message stats) and conversation messages.
func LoadSessionDetail(claudeDir string, session *model.Session) ([]model.Message, error) {
	path := filepath.Join(claudeDir, "projects", session.ProjectPath, session.ID+".jsonl")
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, maxLineBuffer), maxLineBuffer)

	var messages []model.Message
	var lastAssistantEntry Entry

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		// Skip non-message lines without a full JSON parse
		if !isMessageLine(line) {
			continue
		}

		e, err := parseLine(line)
		if err != nil || !e.IsMessage() {
			continue
		}

		switch e.Type {
		case "user":
			session.UserMsgs++
		case "assistant":
			session.AssistantMsgs++
			session.OutputTokens += e.Message.Usage.OutputTokens
			lastAssistantEntry = e
		}

		content := e.MessageContent()
		if content != "" {
			messages = append(messages, model.Message{
				Role:      e.Type,
				Content:   content,
				Timestamp: e.Timestamp,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	session.MessageCount = session.UserMsgs + session.AssistantMsgs

	if lastAssistantEntry.Type == "assistant" {
		u := lastAssistantEntry.Message.Usage
		session.ContextTokens = u.InputTokens + u.CacheCreate + u.CacheReadTokens
		session.CacheTokens = u.CacheReadTokens
		if lastAssistantEntry.Message.Model != "" {
			session.ModelID = lastAssistantEntry.Message.Model
		}
	}

	return messages, nil
}

// isMessageLine checks if a JSONL line is a user or assistant message
// by scanning the first bytes, avoiding a full JSON unmarshal for
// file-history-snapshot, system, and other non-message entries.
func isMessageLine(line []byte) bool {
	// "type" can appear up to ~150 bytes in depending on preceding fields
	n := len(line)
	if n > 200 {
		n = 200
	}
	prefix := line[:n]
	return bytes.Contains(prefix, []byte(`"type":"user"`)) ||
		bytes.Contains(prefix, []byte(`"type":"assistant"`))
}

// LoadRichConversation reads a full JSONL file and returns messages with all
// content blocks (text, images, tool use, tool results, thinking).
func LoadRichConversation(claudeDir string, session model.Session) ([]model.RichMessage, error) {
	path := filepath.Join(claudeDir, "projects", session.ProjectPath, session.ID+".jsonl")

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, maxLineBuffer), maxLineBuffer)

	var messages []model.RichMessage

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		e, err := parseLine(line)
		if err != nil || !e.IsMessage() {
			continue
		}

		blocks := e.RichContentBlocks()
		if len(blocks) == 0 {
			continue
		}

		messages = append(messages, model.RichMessage{
			Role:      e.Type,
			Blocks:    blocks,
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
