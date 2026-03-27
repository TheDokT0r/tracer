package gemini

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"tracer/internal/model"
)

// ScanSessions scans ~/.gemini/tmp/*/chats/ for JSON session files.
func ScanSessions(geminiDir string) ([]model.Session, error) {
	tmpDir := filepath.Join(geminiDir, "tmp")
	var paths []string
	filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}
		// Only look in chats/ directories
		if filepath.Base(filepath.Dir(path)) != "chats" {
			return nil
		}
		paths = append(paths, path)
		return nil
	})

	// Load project name mappings
	projectNames := loadProjectNames(geminiDir)

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

			sess, err := scanSessionFile(path, projectNames)
			if err != nil {
				return
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

// sessionFile is the on-disk format for a Gemini CLI session.
type sessionFile struct {
	SessionID   string           `json:"sessionId"`
	ProjectHash string           `json:"projectHash"`
	StartTime   string           `json:"startTime"`
	LastUpdated string           `json:"lastUpdated"`
	Messages    []geminiMessage  `json:"messages"`
}

type geminiMessage struct {
	ID        string          `json:"id"`
	Timestamp string          `json:"timestamp"`
	Type      string          `json:"type"` // "user" or "gemini"
	Content   json.RawMessage `json:"content"`
	Model     string          `json:"model"`
	Tokens    *geminiTokens   `json:"tokens,omitempty"`
}

type geminiTokens struct {
	Input  int `json:"input"`
	Output int `json:"output"`
	Cached int `json:"cached"`
}

func scanSessionFile(path string, projectNames map[string]string) (model.Session, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return model.Session{}, err
	}

	var sf sessionFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return model.Session{}, err
	}

	sess := model.Session{
		Agent:    model.AgentGemini,
		ID:       sf.SessionID,
		FilePath: path,
	}

	if sf.StartTime != "" {
		sess.StartedAt, _ = time.Parse(time.RFC3339Nano, sf.StartTime)
	}

	// Derive directory from the chats path: .../tmp/<project>/chats/
	// The <project> component maps to a real path via projects.json
	chatsDir := filepath.Dir(path)
	projectDir := filepath.Base(filepath.Dir(chatsDir))
	if realPath, ok := projectNames[projectDir]; ok {
		sess.Directory = realPath
	} else {
		sess.Directory = projectDir
	}

	for _, m := range sf.Messages {
		if m.Type == "user" && sess.Name == "" {
			sess.Name = extractText(m.Content)
		}
		if m.Type == "gemini" && sess.ModelID == "" && m.Model != "" {
			sess.ModelID = m.Model
		}
	}

	if sess.Name == "" {
		return model.Session{}, os.ErrNotExist
	}
	if len([]rune(sess.Name)) > 80 {
		sess.Name = string([]rune(sess.Name)[:80]) + "..."
	}

	return sess, nil
}

// extractText gets the text content from a Gemini message content field.
// Content can be a JSON string or an array of {"text": "..."} blocks.
func extractText(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	// Try as plain string
	var s string
	if json.Unmarshal(raw, &s) == nil {
		s = strings.TrimSpace(s)
		if s != "" && !strings.HasPrefix(s, "/") {
			return s
		}
		return ""
	}
	// Try as array of content blocks
	var blocks []struct {
		Text string `json:"text"`
	}
	if json.Unmarshal(raw, &blocks) == nil {
		for _, b := range blocks {
			text := strings.TrimSpace(b.Text)
			if text != "" && !strings.HasPrefix(text, "/") {
				return text
			}
		}
	}
	return ""
}

// LoadSessionDetail reads the full Gemini session for detail view.
func LoadSessionDetail(path string, session *model.Session) ([]model.Message, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var sf sessionFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return nil, err
	}

	var messages []model.Message
	for _, m := range sf.Messages {
		role := m.Type
		if role == "gemini" {
			role = "assistant"
		}
		if role != "user" && role != "assistant" {
			continue
		}

		text := extractText(m.Content)
		if text == "" {
			continue
		}

		switch role {
		case "user":
			session.UserMsgs++
		case "assistant":
			session.AssistantMsgs++
		}

		if session.ModelID == "" && m.Model != "" {
			session.ModelID = m.Model
		}

		if m.Tokens != nil {
			session.ContextTokens = m.Tokens.Input + m.Tokens.Cached
			session.OutputTokens += m.Tokens.Output
		}

		ts, _ := time.Parse(time.RFC3339Nano, m.Timestamp)
		messages = append(messages, model.Message{
			Role:      role,
			Content:   text,
			Timestamp: ts,
		})
	}

	session.MessageCount = session.UserMsgs + session.AssistantMsgs
	return messages, nil
}

// loadProjectNames reads projects.json to map short names to full paths.
func loadProjectNames(geminiDir string) map[string]string {
	result := make(map[string]string)
	data, err := os.ReadFile(filepath.Join(geminiDir, "projects.json"))
	if err != nil {
		return result
	}
	var pf struct {
		Projects map[string]string `json:"projects"`
	}
	if json.Unmarshal(data, &pf) == nil {
		// Reverse: path -> shortname becomes shortname -> path
		for path, name := range pf.Projects {
			result[name] = path
		}
	}
	return result
}
