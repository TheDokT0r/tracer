package ui

import (
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/viewport"
	"tracer/internal/model"
)

type detailView struct {
	session  model.Session
	messages []model.Message
	viewport viewport.Model
	progress progress.Model
	ready    bool
	width    int
	height   int
}

func newDetailView(session model.Session, messages []model.Message, width, height int) detailView {
	vp := viewport.New(
		viewport.WithWidth(width),
		viewport.WithHeight(height-14),
	)

	prog := progress.New(
		progress.WithWidth(40),
		progress.WithDefaultBlend(),
	)

	d := detailView{
		session:  session,
		messages: messages,
		viewport: vp,
		progress: prog,
		ready:    true,
		width:    width,
		height:   height,
	}

	d.viewport.SetContent(d.conversationContent())
	return d
}

func (d detailView) headerView() string {
	var b strings.Builder

	dir := shortenHome(d.session.Directory)

	pct := d.session.ContextPercent()
	maxTok := d.session.MaxContextTokens()
	inputK := fmt.Sprintf("%dk", d.session.ContextTokens/1000)
	maxK := fmt.Sprintf("%dk", maxTok/1000)
	pctLabel := fmt.Sprintf("%s / %s tokens (%d%%)", inputK, maxK, int(pct*100))
	progressBar := d.progress.ViewAs(pct)
	outputK := fmt.Sprintf("%dk", d.session.OutputTokens/1000)

	b.WriteString(titleStyle.Render(d.session.Name))
	b.WriteString("\n\n")
	b.WriteString(labelStyle.Render("Agent") + valueStyle.Render(d.session.Agent.DisplayName()) + "\n")
	b.WriteString(labelStyle.Render("Session ID") + valueStyle.Render(d.session.ID) + "\n")
	b.WriteString(labelStyle.Render("Date") + valueStyle.Render(d.session.StartedAt.Format("2006-01-02 15:04")) + "\n")
	b.WriteString(labelStyle.Render("Directory") + valueStyle.Render(dir) + "\n")
	if d.session.Branch != "" && d.session.Branch != "-" {
		b.WriteString(labelStyle.Render("Branch") + valueStyle.Render(d.session.Branch) + "\n")
	}
	if d.session.ModelID != "" {
		b.WriteString(labelStyle.Render("Model") + valueStyle.Render(d.session.ModelID) + "\n")
	}
	b.WriteString(labelStyle.Render("Messages") + valueStyle.Render(fmt.Sprintf(
		"%d total (%d user, %d assistant)",
		d.session.MessageCount, d.session.UserMsgs, d.session.AssistantMsgs,
	)) + "\n")
	if d.session.ContextTokens > 0 {
		b.WriteString(labelStyle.Render("Context") + progressBar + " " + valueStyle.Render(pctLabel) + "\n")
	}
	if d.session.OutputTokens > 0 {
		b.WriteString(labelStyle.Render("Output") + valueStyle.Render(outputK+" tokens") + "\n")
	}
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", d.width) + "\n")

	return b.String()
}

func (d detailView) conversationContent() string {
	var b strings.Builder

	assistantLabel := d.session.Agent.DisplayName() + ": "

	for i, msg := range d.messages {
		content := strings.ReplaceAll(msg.Content, "\r\n", "\n")

		switch msg.Role {
		case "user":
			b.WriteString(userStyle.Render("You: "))
			b.WriteString(content)
		case "assistant":
			b.WriteString(assistantStyle.Render(assistantLabel))
			b.WriteString(content)
		}

		if i < len(d.messages)-1 {
			b.WriteString("\n\n")
		}
	}

	return b.String()
}

func (d detailView) view() string {
	if !d.ready {
		return "Loading..."
	}

	header := d.headerView()
	body := d.viewport.View()
	sep := helpSepStyle.Render(" • ")
	help := helpItem("↑/↓", "scroll") + sep +
		helpItem("enter", "resume") + sep +
		helpItem("f", "fork") + sep +
		helpItem("r", "rename") + sep +
		helpItem("e", "edit") + sep +
		helpItem("c", "copy") + sep +
		helpItem("x", "export") + sep +
		helpItem("d", "delete") + sep +
		helpItem("esc", "back")

	return header + body + "\n" + help
}

// shortenHome replaces the user's home directory prefix with ~.
func shortenHome(path string) string {
	home := os.Getenv("HOME")
	if home != "" && strings.HasPrefix(path, home) {
		return "~" + path[len(home):]
	}
	return path
}
