package claude

import (
	"bufio"
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

		// Pick up cwd/branch/timestamp from first entry that has them
		if sess.Directory == "" && e.CWD != "" {
			sess.Directory = e.CWD
		}
		if sess.Branch == "" && e.GitBranch != "" {
			sess.Branch = e.GitBranch
		}
		if sess.StartedAt.IsZero() && !e.Timestamp.IsZero() {
			sess.StartedAt = e.Timestamp
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

// LoadSessionDetails reads the full JSONL file to populate token counts and message stats.
// Called when opening the detail view.
func LoadSessionDetails(claudeDir string, session *model.Session) {
	path := filepath.Join(claudeDir, "projects", session.ProjectPath, session.ID+".jsonl")
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, maxLineBuffer), maxLineBuffer)

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
			session.UserMsgs++
		case "assistant":
			session.AssistantMsgs++
			session.OutputTokens += e.Message.Usage.OutputTokens
			lastAssistantEntry = e
		}
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
