package codex

import (
	"os"
	"os/exec"
	"tracer/internal/model"
)

// Provider implements model.Provider for Codex CLI.
type Provider struct {
	dir string
}

func NewProvider(dir string) *Provider {
	return &Provider{dir: dir}
}

func (p *Provider) Type() model.Agent { return model.AgentCodex }

func (p *Provider) Scan() ([]model.Session, error) {
	return ScanSessions(p.dir)
}

func (p *Provider) LoadDetail(session *model.Session) ([]model.Message, error) {
	return LoadSessionDetail(session.FilePath, session)
}

func (p *Provider) LoadRichMessages(_ model.Session) ([]model.RichMessage, error) {
	return nil, nil
}

func (p *Provider) ResumeArgs(session model.Session) (string, []string, bool) {
	bin, err := exec.LookPath("codex")
	if err != nil {
		return "", nil, false
	}
	return bin, []string{"resume", session.ID}, true
}

func (p *Provider) ForkArgs(session model.Session) (string, []string, bool) {
	bin, err := exec.LookPath("codex")
	if err != nil {
		return "", nil, false
	}
	return bin, []string{"fork", session.ID}, true
}

func (p *Provider) NewArgs(_ string) (string, []string) {
	bin, err := exec.LookPath("codex")
	if err != nil {
		return "", nil
	}
	return bin, nil
}

func (p *Provider) DeleteSession(session model.Session) error {
	if session.FilePath != "" {
		return os.Remove(session.FilePath)
	}
	return model.ErrUnsupported
}

func (p *Provider) WriteRename(_ model.Session, _ string) {}
