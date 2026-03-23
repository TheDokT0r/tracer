package claude

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"tracer/internal/model"
)

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

// ExportMarkdown writes the session conversation as a Markdown file.
func ExportMarkdown(session model.Session, messages []model.Message) (string, error) {
	var b strings.Builder

	b.WriteString("# " + session.Name + "\n\n")
	writeMetadataTable(&b, session)
	b.WriteString("\n---\n\n")

	for i, msg := range messages {
		switch msg.Role {
		case "user":
			b.WriteString("## You\n\n")
		case "assistant":
			b.WriteString("## Claude\n\n")
		}
		b.WriteString(msg.Content)
		if i < len(messages)-1 {
			b.WriteString("\n\n---\n\n")
		} else {
			b.WriteString("\n")
		}
	}

	return writeExportFile(session, b.String(), "md")
}

// ExportHTML writes the session conversation as an HTML file with
// embedded images, collapsible long messages, and syntax highlighting.
func ExportHTML(session model.Session, messages []model.RichMessage) (string, error) {
	var b strings.Builder

	title := html.EscapeString(session.Name)

	b.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>` + title + `</title>
<style>` + htmlCSS + `</style>
</head>
<body>
`)

	// Sticky header with metadata
	b.WriteString(`<div class="header"><div class="header-inner">`)
	b.WriteString(`<h1>` + title + `</h1>`)
	writeHTMLMetadata(&b, session)
	b.WriteString(`</div></div>`)

	// Chat area
	b.WriteString(`<div class="chat">`)
	for i, msg := range messages {
		writeHTMLMessage(&b, msg, i)
	}
	b.WriteString(`</div>`)

	b.WriteString(`<script>` + htmlJS + `</script>
</body>
</html>
`)

	return writeExportFile(session, b.String(), "html")
}

func writeMetadataTable(b *strings.Builder, s model.Session) {
	b.WriteString("| Field | Value |\n")
	b.WriteString("|-------|-------|\n")
	b.WriteString(fmt.Sprintf("| Session ID | `%s` |\n", s.ID))
	b.WriteString(fmt.Sprintf("| Date | %s |\n", s.StartedAt.Format("2006-01-02 15:04")))
	b.WriteString(fmt.Sprintf("| Directory | %s |\n", s.Directory))
	if s.Branch != "" {
		b.WriteString(fmt.Sprintf("| Branch | %s |\n", s.Branch))
	}
	if s.ModelID != "" {
		b.WriteString(fmt.Sprintf("| Model | %s |\n", s.ModelID))
	}
	b.WriteString(fmt.Sprintf("| Messages | %d total (%d user, %d assistant) |\n",
		s.MessageCount, s.UserMsgs, s.AssistantMsgs))
	if s.ContextTokens > 0 {
		b.WriteString(fmt.Sprintf("| Context | %dk tokens |\n", s.ContextTokens/1000))
	}
	if s.OutputTokens > 0 {
		b.WriteString(fmt.Sprintf("| Output | %dk tokens |\n", s.OutputTokens/1000))
	}
}

func writeHTMLMetadata(b *strings.Builder, s model.Session) {
	b.WriteString(`<div class="meta-row">`)
	item := func(label, value string) {
		b.WriteString(`<span><span class="meta-label">` + label + `</span> ` + html.EscapeString(value) + `</span>`)
	}
	item("", s.StartedAt.Format("Jan 2, 2006 15:04"))
	item("·", s.Directory)
	if s.Branch != "" && s.Branch != "-" {
		item("·", s.Branch)
	}
	if s.ModelID != "" {
		item("·", s.ModelID)
	}
	item("·", fmt.Sprintf("%d messages", s.MessageCount))
	if s.ContextTokens > 0 {
		item("·", fmt.Sprintf("%dk ctx", s.ContextTokens/1000))
	}
	b.WriteString(`</div>`)
}

const collapseThreshold = 2000 // characters

func writeHTMLMessage(b *strings.Builder, msg model.RichMessage, index int) {
	role := "user"
	name := "You"
	avatarClass := "avatar-user"
	avatarContent := userInitial
	if msg.Role == "assistant" {
		role = "assistant"
		name = "Claude"
		avatarClass = "avatar-claude"
		avatarContent = claudeSVG
	}

	fullText := msg.Text()
	needsCollapse := len([]rune(fullText)) > collapseThreshold

	b.WriteString(fmt.Sprintf(`<div class="msg %s">`, role))

	// Name row: avatar + name + timestamp
	b.WriteString(`<div class="msg-name"><div class="` + avatarClass + `">` + avatarContent + `</div>`)
	b.WriteString(`<span class="name">` + name + `</span>`)
	if !msg.Timestamp.IsZero() {
		b.WriteString(fmt.Sprintf(`<span class="ts">%s</span>`, msg.Timestamp.Format("15:04")))
	}
	b.WriteString(`</div>`)

	// Content
	if needsCollapse {
		b.WriteString(fmt.Sprintf(`<div class="msg-content collapsible" id="msg-%d">`, index))
	} else {
		b.WriteString(`<div class="msg-content">`)
	}

	for _, block := range msg.Blocks {
		switch block.Type {
		case "text":
			b.WriteString(`<div class="text-block">` + html.EscapeString(block.Text) + `</div>`)
		case "image":
			if block.Data != "" {
				mediaType := block.MediaType
				if mediaType == "" {
					mediaType = "image/png"
				}
				b.WriteString(fmt.Sprintf(`<div class="img-block"><img src="data:%s;base64,%s" alt="image"></div>`, mediaType, block.Data))
			}
		case "thinking":
			b.WriteString(`<details class="detail-block thinking"><summary>Thinking</summary><pre>` + html.EscapeString(block.Text) + `</pre></details>`)
		case "tool_use":
			b.WriteString(`<details class="detail-block tool-use"><summary>` + html.EscapeString(block.ToolName) + `</summary>`)
			if block.ToolInput != "" {
				b.WriteString(`<pre>` + html.EscapeString(block.ToolInput) + `</pre>`)
			}
			b.WriteString(`</details>`)
		case "tool_result":
			if block.Text != "" {
				b.WriteString(`<details class="detail-block tool-result"><summary>Result</summary><pre>` + html.EscapeString(block.Text) + `</pre></details>`)
			}
		}
	}

	b.WriteString(`</div>`) // msg-content

	if needsCollapse {
		b.WriteString(fmt.Sprintf(`<button class="show-more" onclick="toggleCollapse('msg-%d', this)">Show more</button>`, index))
	}

	b.WriteString(`</div>`) // msg
}

const claudeSVG = `<svg width="16" height="16" viewBox="0 0 24 24" fill="none"><path d="M16.862 11.487l-4.235-8.136a.667.667 0 0 0-1.254 0L7.138 11.487a.667.667 0 0 0 .362.893l4.167 1.632a.667.667 0 0 0 .666 0l4.167-1.632a.667.667 0 0 0 .362-.893z" fill="currentColor"/><path d="M13.87 16.394l-1.535-.601a.667.667 0 0 0-.67 0l-1.535.601a.667.667 0 0 0-.362.893l1.535 3.577a.667.667 0 0 0 1.254 0l1.535-3.577a.667.667 0 0 0-.222-.893z" fill="currentColor"/></svg>`

const userInitial = `<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>`

func writeExportFile(session model.Session, content, ext string) (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	exportDir := filepath.Join(configDir, "tracer", "exports")
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return "", err
	}

	slug := slugify(session.Name)
	idPrefix := session.ID
	if len(idPrefix) > 8 {
		idPrefix = idPrefix[:8]
	}
	filename := fmt.Sprintf("%s-%s.%s", slug, idPrefix, ext)
	path := filepath.Join(exportDir, filename)

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return "", err
	}
	return path, nil
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = slugRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 50 {
		s = s[:50]
		s = strings.TrimRight(s, "-")
	}
	if s == "" {
		s = "session"
	}
	return s
}

const htmlCSS = `
*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

body {
    font-family: "Söhne", -apple-system, BlinkMacSystemFont, "Segoe UI", "Helvetica Neue", sans-serif;
    line-height: 1.7;
    color: #ececec;
    background: #2b2b2b;
    -webkit-font-smoothing: antialiased;
}

/* ---- Sticky header ---- */
.header {
    background: rgba(43,43,43,0.85);
    backdrop-filter: blur(12px);
    -webkit-backdrop-filter: blur(12px);
    border-bottom: 1px solid rgba(255,255,255,0.06);
    padding: 0.9rem 0;
    position: sticky;
    top: 0;
    z-index: 100;
}
.header-inner { max-width: 680px; margin: 0 auto; padding: 0 1.5rem; }
.header h1 {
    font-size: 0.95rem;
    font-weight: 500;
    color: #e8e8e8;
    letter-spacing: -0.01em;
}
.meta-row {
    display: flex;
    flex-wrap: wrap;
    gap: 0.15rem 0.6rem;
    font-size: 0.72rem;
    color: #888;
    margin-top: 0.3rem;
}
.meta-row span { white-space: nowrap; }

/* ---- Chat container ---- */
.chat {
    max-width: 680px;
    margin: 0 auto;
    padding: 0.5rem 1.5rem 5rem;
}

/* ---- Single message ---- */
.msg { padding: 1.5rem 0; }
.msg + .msg { border-top: 1px solid rgba(255,255,255,0.04); }

/* User messages: rounded panel like claude.ai */
.msg.user .msg-content {
    background: #303030;
    border-radius: 1.25rem;
    padding: 0.9rem 1.15rem;
    margin-top: 0.4rem;
}

/* Claude messages: no panel, just text */
.msg.assistant .msg-content {
    margin-top: 0.4rem;
}

/* ---- Name row ---- */
.msg-name {
    display: flex;
    align-items: center;
    gap: 0.5rem;
}
.name {
    font-size: 0.82rem;
    font-weight: 600;
    color: #e8e8e8;
}
.ts {
    font-size: 0.7rem;
    color: #666;
    margin-left: auto;
}

/* ---- Avatars ---- */
.avatar-claude, .avatar-user {
    width: 24px;
    height: 24px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
}
.avatar-claude {
    background: #d97757;
    color: #fff;
}
.avatar-user {
    background: #6b7280;
    color: #fff;
}

/* ---- Content / Collapse ---- */
.msg-content { position: relative; }
.msg-content.collapsible {
    max-height: 420px;
    overflow: hidden;
}
.msg-content.collapsible::after {
    content: "";
    position: absolute;
    bottom: 0; left: 0; right: 0;
    height: 90px;
    pointer-events: none;
}
/* Gradient matches the background behind each role */
.msg.user .msg-content.collapsible::after {
    background: linear-gradient(transparent, #303030);
    border-radius: 0 0 1.25rem 1.25rem;
}
.msg.assistant .msg-content.collapsible::after {
    background: linear-gradient(transparent, #2b2b2b);
}
.msg-content.expanded { max-height: none; }
.msg-content.expanded::after { display: none; }

.show-more {
    display: inline-block;
    margin-top: 0.6rem;
    background: none;
    border: 1px solid #444;
    border-radius: 9999px;
    color: #aaa;
    cursor: pointer;
    font-size: 0.75rem;
    padding: 0.25rem 1rem;
    transition: all 0.15s;
}
.show-more:hover { border-color: #666; color: #ddd; }

/* ---- Text ---- */
.text-block {
    white-space: pre-wrap;
    word-wrap: break-word;
    font-size: 0.925rem;
    line-height: 1.7;
    color: #e8e8e8;
}

/* ---- Images ---- */
.img-block { margin: 0.75rem 0; }
.img-block img {
    max-width: 100%;
    max-height: 500px;
    border-radius: 12px;
}

/* ---- Details (thinking, tools) ---- */
.detail-block {
    margin: 0.6rem 0;
    border: 1px solid rgba(255,255,255,0.08);
    border-radius: 12px;
    overflow: hidden;
}
.detail-block summary {
    padding: 0.5rem 0.9rem;
    cursor: pointer;
    color: #999;
    font-size: 0.78rem;
    user-select: none;
    list-style: none;
    display: flex;
    align-items: center;
    gap: 0.35rem;
}
.detail-block summary::-webkit-details-marker { display: none; }
.detail-block summary::before {
    content: "▸";
    font-size: 0.7rem;
    transition: transform 0.15s;
}
.detail-block[open] summary::before { transform: rotate(90deg); }
.detail-block summary:hover { color: #ccc; }
.detail-block pre {
    margin: 0;
    padding: 0.75rem 0.9rem;
    font-size: 0.78rem;
    font-family: "SF Mono", "Fira Code", "JetBrains Mono", Menlo, monospace;
    overflow-x: auto;
    white-space: pre-wrap;
    word-wrap: break-word;
    color: #aaa;
    background: rgba(0,0,0,0.15);
    line-height: 1.55;
}
.thinking summary { color: #d4a253; }
.tool-use summary { color: #6cc070; }
.tool-result summary { color: #6ba3d6; }

@media (max-width: 640px) {
    .header-inner, .chat { padding-left: 1rem; padding-right: 1rem; }
}
`

const htmlJS = `
function toggleCollapse(id, btn) {
    var el = document.getElementById(id);
    el.classList.toggle("expanded");
    btn.textContent = el.classList.contains("expanded") ? "Show less" : "Show more";
}
`
