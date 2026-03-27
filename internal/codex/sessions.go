package codex

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"tracer/internal/model"
)

const maxLineBuffer = 1024 * 1024

// ScanSessions scans ~/.codex/sessions/ for JSONL session files.
func ScanSessions(codexDir string) ([]model.Session, error) {
	sessionsDir := filepath.Join(codexDir, "sessions")
	var paths []string
	filepath.Walk(sessionsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".jsonl") {
			return nil
		}
		paths = append(paths, path)
		return nil
	})

	// Load thread names from session_index.jsonl
	threadNames := loadThreadNames(codexDir)

	var mu sync.Mutex
	var wg sync.WaitGroup
	sessions := make([]model.Session, 0, len(paths))

	sem := make(chan struct{}, 16)
	for _, p := range paths {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			sess, err := scanSessionHead(path)
			if err != nil {
				return
			}
			if name, ok := threadNames[sess.ID]; ok {
				sess.Name = name
			}

			mu.Lock()
			sessions = append(sessions, sess)
			mu.Unlock()
		}(p)
	}
	wg.Wait()

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].StartedAt.After(sessions[j].StartedAt)
	})

	return sessions, nil
}

// scanSessionHead reads a Codex JSONL file until it finds the session meta
// and first user message.
func scanSessionHead(path string) (model.Session, error) {
	f, err := os.Open(path)
	if err != nil {
		return model.Session{}, err
	}
	defer f.Close()

	sess := model.Session{
		Agent:    model.AgentCodex,
		FilePath: path,
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, maxLineBuffer), maxLineBuffer)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry struct {
			Type      string    `json:"type"`
			Timestamp time.Time `json:"timestamp"`
			Payload   json.RawMessage `json:"payload"`
		}
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}

		switch entry.Type {
		case "session_meta":
			var meta struct {
				ID  string `json:"id"`
				CWD string `json:"cwd"`
			}
			json.Unmarshal(entry.Payload, &meta)
			sess.ID = meta.ID
			sess.Directory = meta.CWD
			if sess.StartedAt.IsZero() && !entry.Timestamp.IsZero() {
				sess.StartedAt = entry.Timestamp
			}

		case "turn_context":
			var ctx struct {
				Model string `json:"model"`
				CWD   string `json:"cwd"`
			}
			json.Unmarshal(entry.Payload, &ctx)
			if sess.ModelID == "" && ctx.Model != "" {
				sess.ModelID = ctx.Model
			}
			if sess.Directory == "" && ctx.CWD != "" {
				sess.Directory = ctx.CWD
			}

		case "response_item":
			var item struct {
				Role    string `json:"role"`
				Content []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				} `json:"content"`
			}
			json.Unmarshal(entry.Payload, &item)

			if item.Role == "user" && sess.Name == "" {
				for _, c := range item.Content {
					text := strings.TrimSpace(c.Text)
					// Skip system/permission messages
					if text == "" || strings.HasPrefix(text, "<") {
						continue
					}
					if len([]rune(text)) > 80 {
						text = string([]rune(text)[:80]) + "..."
					}
					sess.Name = text
					break
				}
				if sess.Name != "" {
					return sess, nil
				}
			}
		}
	}

	if sess.ID == "" || sess.Name == "" {
		return model.Session{}, os.ErrNotExist
	}
	return sess, nil
}

// LoadSessionDetail reads the full Codex JSONL file for detail view.
func LoadSessionDetail(path string, session *model.Session) ([]model.Message, error) {
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

		var entry struct {
			Type      string          `json:"type"`
			Timestamp time.Time       `json:"timestamp"`
			Payload   json.RawMessage `json:"payload"`
		}
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}

		if entry.Type != "response_item" {
			continue
		}

		var item struct {
			Role    string `json:"role"`
			Content []struct {
				Type  string `json:"type"`
				Text  string `json:"text"`
				Input string `json:"input"`
			} `json:"content"`
		}
		if err := json.Unmarshal(entry.Payload, &item); err != nil {
			continue
		}

		if item.Role != "user" && item.Role != "assistant" {
			continue
		}

		var text string
		for _, c := range item.Content {
			if c.Type == "text" && c.Text != "" {
				text = c.Text
				break
			}
			if c.Type == "input_text" && c.Text != "" {
				// Skip system messages
				if strings.HasPrefix(c.Text, "<") {
					continue
				}
				text = c.Text
				break
			}
		}
		if text == "" {
			continue
		}

		switch item.Role {
		case "user":
			session.UserMsgs++
		case "assistant":
			session.AssistantMsgs++
		}

		messages = append(messages, model.Message{
			Role:      item.Role,
			Content:   text,
			Timestamp: entry.Timestamp,
		})
	}

	session.MessageCount = session.UserMsgs + session.AssistantMsgs
	return messages, scanner.Err()
}

// loadThreadNames reads session_index.jsonl for named sessions.
func loadThreadNames(codexDir string) map[string]string {
	names := make(map[string]string)
	f, err := os.Open(filepath.Join(codexDir, "session_index.jsonl"))
	if err != nil {
		return names
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var entry struct {
			ID         string `json:"id"`
			ThreadName string `json:"thread_name"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &entry); err == nil && entry.ThreadName != "" {
			names[entry.ID] = entry.ThreadName
		}
	}
	return names
}
