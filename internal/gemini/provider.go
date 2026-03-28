package gemini

import (
	"os/exec"
	"tracer/internal/model"
)

// Provider implements model.Provider for Gemini CLI.
type Provider struct {
	dir string
}

func NewProvider(dir string) *Provider {
	return &Provider{dir: dir}
}

func (p *Provider) Type() model.Agent { return model.AgentGemini }

func (p *Provider) Scan() ([]model.Session, error) {
	return ScanSessions(p.dir)
}

func (p *Provider) LoadDetail(session *model.Session) ([]model.Message, error) {
	return LoadSessionDetail(session.FilePath, session)
}

func (p *Provider) LoadRichMessages(_ model.Session) ([]model.RichMessage, error) {
	return nil, nil
}

func (p *Provider) ResumeArgs(_ model.Session) (string, []string, bool) {
	return "", nil, false
}

func (p *Provider) ForkArgs(_ model.Session) (string, []string, bool) {
	return "", nil, false
}

func (p *Provider) NewArgs(_ string) (string, []string) {
	bin, err := exec.LookPath("gemini")
	if err != nil {
		return "", nil
	}
	return bin, nil
}

func (p *Provider) DeleteSession(_ model.Session) error {
	return model.ErrUnsupported
}

func (p *Provider) WriteRename(_ model.Session, _ string) {}
