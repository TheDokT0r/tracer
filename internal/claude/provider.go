package claude

import (
	"os/exec"
	"tracer/internal/model"
)

// Provider implements model.Provider for Claude Code.
type Provider struct {
	dir   string
	model string // optional model override from config
}

func NewProvider(dir, modelOverride string) *Provider {
	return &Provider{dir: dir, model: modelOverride}
}

func (p *Provider) Type() model.Agent { return model.AgentClaude }

func (p *Provider) Scan() ([]model.Session, error) {
	return ScanSessions(p.dir)
}

func (p *Provider) LoadDetail(session *model.Session) ([]model.Message, error) {
	return LoadSessionDetail(p.dir, session)
}

func (p *Provider) LoadRichMessages(session model.Session) ([]model.RichMessage, error) {
	return LoadRichConversation(p.dir, session)
}

func (p *Provider) ResumeArgs(session model.Session) (string, []string, bool) {
	bin, err := exec.LookPath("claude")
	if err != nil {
		return "", nil, false
	}
	args := []string{"--resume", session.ID}
	if p.model != "" {
		args = append(args, "--model", p.model)
	}
	if session.Name != "" {
		args = append(args, "--name", session.Name)
	}
	return bin, args, true
}

func (p *Provider) ForkArgs(session model.Session) (string, []string, bool) {
	bin, err := exec.LookPath("claude")
	if err != nil {
		return "", nil, false
	}
	args := []string{"--resume", session.ID, "--fork-session"}
	if p.model != "" {
		args = append(args, "--model", p.model)
	}
	if session.Name != "" {
		args = append(args, "--name", session.Name)
	}
	return bin, args, true
}

func (p *Provider) NewArgs(dir string) (string, []string) {
	bin, err := exec.LookPath("claude")
	if err != nil {
		return "", nil
	}
	var args []string
	if p.model != "" {
		args = append(args, "--model", p.model)
	}
	return bin, args
}

func (p *Provider) DeleteSession(session model.Session) error {
	return DeleteSession(p.dir, session)
}

func (p *Provider) WriteRename(session model.Session, name string) {
	WriteRename(p.dir, session, name)
}
