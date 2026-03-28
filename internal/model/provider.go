package model

import "errors"

// ErrUnsupported indicates an operation is not supported by this agent.
var ErrUnsupported = errors.New("not supported")

// Provider defines the interface that each AI agent must implement.
// Methods that return (value, error) use ErrUnsupported to signal
// that the operation is not available for this agent.
type Provider interface {
	// Type returns the agent identifier.
	Type() Agent

	// Scan discovers all sessions for this agent.
	Scan() ([]Session, error)

	// LoadDetail reads the full session file, populating metadata
	// and returning conversation messages.
	LoadDetail(session *Session) ([]Message, error)

	// LoadRichMessages returns rich content blocks (images, tool use, thinking).
	// Returns nil, nil if not supported.
	LoadRichMessages(session Session) ([]RichMessage, error)

	// ResumeArgs returns the CLI binary and arguments to resume a session.
	// ok=false means resume is not supported.
	ResumeArgs(session Session) (bin string, args []string, ok bool)

	// ForkArgs returns the CLI binary and arguments to fork a session.
	// ok=false means fork is not supported.
	ForkArgs(session Session) (bin string, args []string, ok bool)

	// NewArgs returns the CLI binary and arguments to start a new session.
	NewArgs(dir string) (bin string, args []string)

	// DeleteSession removes all files associated with a session.
	// Returns ErrUnsupported if deletion is not available.
	DeleteSession(session Session) error

	// WriteRename persists a rename in the agent's native format.
	// No-op if the agent doesn't support renames.
	WriteRename(session Session, name string)
}
