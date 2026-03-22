package claude

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"tracer/internal/model"
)

// DeleteSession removes all files associated with a session:
// the JSONL file, subdirectory, file-history, tasks, and history.jsonl entries.
func DeleteSession(claudeDir string, session model.Session) error {
	// 1. Session JSONL file
	jsonlPath := filepath.Join(claudeDir, "projects", session.ProjectPath, session.ID+".jsonl")
	if err := removeIgnoreNotFound(os.Remove(jsonlPath)); err != nil {
		return err
	}

	// 2. Session subdirectory (subagents, tool-results)
	subDir := filepath.Join(claudeDir, "projects", session.ProjectPath, session.ID)
	if err := removeIgnoreNotFound(os.RemoveAll(subDir)); err != nil {
		return err
	}

	// 3. File version history
	fileHistDir := filepath.Join(claudeDir, "file-history", session.ID)
	if err := removeIgnoreNotFound(os.RemoveAll(fileHistDir)); err != nil {
		return err
	}

	// 4. Task definitions
	tasksDir := filepath.Join(claudeDir, "tasks", session.ID)
	if err := removeIgnoreNotFound(os.RemoveAll(tasksDir)); err != nil {
		return err
	}

	// 5. Entries from history.jsonl
	if err := removeFromHistory(claudeDir, session.ID); err != nil {
		return err
	}

	return nil
}

// removeFromHistory filters out lines from history.jsonl that match the given session ID.
func removeFromHistory(claudeDir, sessionID string) error {
	histPath := filepath.Join(claudeDir, "history.jsonl")

	f, err := os.Open(histPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer f.Close()

	var kept [][]byte
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, maxLineBuffer), maxLineBuffer)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry struct {
			SessionID string `json:"sessionId"`
		}
		if err := json.Unmarshal(line, &entry); err != nil {
			// Keep unparseable lines
			kept = append(kept, bytes.Clone(line))
			continue
		}

		if entry.SessionID == sessionID {
			continue
		}

		kept = append(kept, bytes.Clone(line))
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	f.Close()

	var buf bytes.Buffer
	for _, line := range kept {
		buf.Write(line)
		buf.WriteByte('\n')
	}

	return os.WriteFile(histPath, buf.Bytes(), 0o644)
}

// removeIgnoreNotFound returns nil if the error is a "not found" error.
func removeIgnoreNotFound(err error) error {
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
