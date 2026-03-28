package export

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

	agentName := session.Agent.DisplayName()
	for i, msg := range messages {
		switch msg.Role {
		case "user":
			b.WriteString("## You\n\n")
		case "assistant":
			b.WriteString("## " + agentName + "\n\n")
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

	// Chat area with inline header
	b.WriteString(`<div class="chat">`)
	b.WriteString(`<div class="chat-header">`)
	b.WriteString(`<h1>` + title + `</h1>`)
	writeHTMLMetadata(&b, session)
	b.WriteString(`</div>`)

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
	b.WriteString(fmt.Sprintf("| Agent | %s |\n", s.Agent.DisplayName()))
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
	item("", s.Agent.DisplayName())
	item("·", s.StartedAt.Format("Jan 2, 2006 15:04"))
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
	role := msg.Role
	if role == "" {
		role = "user"
	}

	fullText := msg.Text()
	needsCollapse := len([]rune(fullText)) > collapseThreshold

	b.WriteString(fmt.Sprintf(`<div class="msg %s">`, role))
	b.WriteString(`<div class="bubble">`)

	// Content wrapper
	if needsCollapse {
		b.WriteString(fmt.Sprintf(`<div class="msg-content collapsible" id="msg-%d">`, index))
	} else {
		b.WriteString(`<div class="msg-content">`)
	}

	for _, block := range msg.Blocks {
		switch block.Type {
		case "text":
			if role == "assistant" {
				b.WriteString(`<div class="text-block">` + renderMarkdown(block.Text) + `</div>`)
			} else {
				b.WriteString(`<div class="text-block user-text">` + html.EscapeString(block.Text) + `</div>`)
			}
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

	// Timestamp
	if !msg.Timestamp.IsZero() {
		b.WriteString(fmt.Sprintf(`<div class="ts">%s</div>`, msg.Timestamp.Format("15:04")))
	}

	b.WriteString(`</div>`) // bubble
	b.WriteString(`</div>`) // msg
}

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
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "Helvetica Neue", sans-serif;
    line-height: 1.6;
    color: #d4d4d4;
    background: #0b141a;
    -webkit-font-smoothing: antialiased;
}

/* ---- Chat container ---- */
.chat {
    max-width: 100%;
    padding: 1rem 4% 5rem;
}

/* ---- Chat header ---- */
.chat-header {
    text-align: center;
    margin-bottom: 1.5rem;
    padding-bottom: 1rem;
}
.chat-header h1 {
    font-size: 1rem;
    font-weight: 500;
    color: #e8e8e8;
}
.meta-row {
    display: flex;
    flex-wrap: wrap;
    justify-content: center;
    gap: 0.15rem 0.6rem;
    font-size: 0.7rem;
    color: #667781;
    margin-top: 0.3rem;
}
.meta-row span { white-space: nowrap; }

/* ---- Message row ---- */
.msg {
    display: flex;
    padding: 0.15rem 0;
}
.msg.user { justify-content: flex-end; }
.msg.assistant { justify-content: flex-start; }

/* ---- Bubble ---- */
.bubble {
    max-width: 75%;
    padding: 0.5rem 0.75rem;
    border-radius: 12px;
    position: relative;
}
.msg.user .bubble {
    background: #005c4b;
    border-top-right-radius: 4px;
}
.msg.assistant .bubble {
    background: #202c33;
    border-top-left-radius: 4px;
}

/* ---- Timestamp inside bubble ---- */
.ts {
    font-size: 0.65rem;
    color: rgba(255,255,255,0.45);
    text-align: right;
    margin-top: 0.25rem;
}

/* ---- Content / Collapse ---- */
.msg-content { position: relative; }
.msg-content.collapsible {
    max-height: 600px;
    overflow: hidden;
}
.msg-content.collapsible::after {
    content: "";
    position: absolute;
    bottom: 0; left: 0; right: 0;
    height: 90px;
    pointer-events: none;
}
.msg.user .msg-content.collapsible::after {
    background: linear-gradient(transparent, #005c4b);
}
.msg.assistant .msg-content.collapsible::after {
    background: linear-gradient(transparent, #202c33);
}
.msg-content.expanded { max-height: none; }
.msg-content.expanded::after { display: none; }

.show-more {
    display: inline-block;
    margin-top: 0.4rem;
    background: none;
    border: 1px solid rgba(255,255,255,0.2);
    border-radius: 9999px;
    color: #aaa;
    cursor: pointer;
    font-size: 0.7rem;
    padding: 0.2rem 0.8rem;
    transition: all 0.15s;
}
.show-more:hover { border-color: rgba(255,255,255,0.4); color: #ddd; }

/* ---- Text blocks ---- */
.text-block {
    font-size: 0.9rem;
    line-height: 1.55;
    color: #e4e6eb;
}
.text-block.user-text {
    white-space: pre-wrap;
    word-wrap: break-word;
}

/* ---- Markdown elements ---- */
.text-block p { margin: 0.5em 0; }
.text-block p:first-child { margin-top: 0; }
.text-block p:last-child { margin-bottom: 0; }

.text-block h1, .text-block h2, .text-block h3,
.text-block h4, .text-block h5, .text-block h6 {
    color: #e8e8e8;
    margin: 1em 0 0.3em;
    line-height: 1.3;
}
.text-block h1 { font-size: 1.25em; font-weight: 700; }
.text-block h2 { font-size: 1.12em; font-weight: 700; }
.text-block h3 { font-size: 1em; font-weight: 600; }
.text-block h4, .text-block h5, .text-block h6 { font-size: 0.95em; font-weight: 600; }

.text-block strong { font-weight: 600; color: #e8e8e8; }

.text-block a { color: #53bdeb; text-decoration: none; }
.text-block a:hover { text-decoration: underline; }

.text-block code {
    background: rgba(255,255,255,0.08);
    padding: 0.12em 0.3em;
    border-radius: 4px;
    font-size: 0.85em;
    font-family: "SF Mono", "Fira Code", Menlo, monospace;
    color: #e0e0e0;
}

/* ---- Code blocks with copy button ---- */
.code-wrapper {
    position: relative;
    margin: 0.6em 0;
}
.copy-btn {
    position: absolute;
    top: 0.5rem;
    right: 0.5rem;
    background: rgba(255,255,255,0.1);
    border: none;
    border-radius: 6px;
    color: #aaa;
    cursor: pointer;
    padding: 0.25rem 0.5rem;
    font-size: 0.7rem;
    font-family: inherit;
    transition: all 0.15s;
    z-index: 1;
}
.copy-btn:hover { background: rgba(255,255,255,0.2); color: #fff; }
.copy-btn.copied { color: #4ade80; }

.text-block pre {
    background: #0d1117;
    border-radius: 10px;
    padding: 1em 1.2em;
    padding-top: 2.2em;
    overflow-x: auto;
    margin: 0;
    line-height: 1.5;
    font-size: 0.82em;
    font-family: "SF Mono", "Fira Code", Menlo, monospace;
}
.text-block pre code {
    background: none;
    padding: 0;
    border-radius: 0;
    font-size: inherit;
    color: inherit;
}

.text-block blockquote {
    border-left: 3px solid #3b4a54;
    padding-left: 0.8em;
    color: #8696a0;
    margin: 0.5em 0;
}

.text-block ul, .text-block ol {
    padding-left: 1.4em;
    margin: 0.4em 0;
}
.text-block li { margin: 0.15em 0; }

.text-block table {
    border-collapse: collapse;
    margin: 0.6em 0;
    font-size: 0.88em;
    width: auto;
}
.text-block th, .text-block td {
    padding: 0.4em 0.8em;
    text-align: left;
}
.text-block th {
    font-weight: 600;
    color: #e0e0e0;
    border-bottom: 2px solid #3b4a54;
}
.text-block td {
    border-bottom: 1px solid #2a3942;
}

.text-block hr {
    border: none;
    border-top: 1px solid #2a3942;
    margin: 1em 0;
}

.text-block img {
    max-width: 100%;
    border-radius: 8px;
}

/* ---- Images (base64 embedded) ---- */
.img-block { margin: 0.5rem 0; }
.img-block img {
    max-width: 100%;
    max-height: 500px;
    border-radius: 10px;
}

/* ---- Details (thinking, tools) ---- */
.detail-block {
    margin: 0.4rem 0;
    background: rgba(0,0,0,0.2);
    border-radius: 8px;
    overflow: hidden;
}
.detail-block summary {
    padding: 0.4rem 0.7rem;
    cursor: pointer;
    font-size: 0.78rem;
    user-select: none;
    list-style: none;
    display: flex;
    align-items: center;
    gap: 0.35rem;
    color: #8696a0;
}
.detail-block summary::-webkit-details-marker { display: none; }
.detail-block summary::before {
    content: "›";
    font-size: 0.85rem;
    font-weight: 600;
    transition: transform 0.15s;
    display: inline-block;
}
.detail-block[open] summary::before { transform: rotate(90deg); }
.detail-block summary:hover { color: #aebac1; }
.detail-block pre {
    margin: 0;
    padding: 0.6rem 0.7rem;
    font-size: 0.75rem;
    font-family: "SF Mono", "Fira Code", Menlo, monospace;
    overflow-x: auto;
    white-space: pre-wrap;
    word-wrap: break-word;
    color: #8696a0;
    background: rgba(0,0,0,0.15);
    line-height: 1.5;
}
.thinking summary { color: #d4a253; }
.tool-use summary { color: #8bc38e; }
.tool-result summary { color: #53bdeb; }

@media (max-width: 640px) {
    .chat { padding-left: 2%; padding-right: 2%; }
    .bubble { max-width: 90%; }
}
`

const htmlJS = `
function toggleCollapse(id, btn) {
    var el = document.getElementById(id);
    el.classList.toggle("expanded");
    btn.textContent = el.classList.contains("expanded") ? "Show less" : "Show more";
}

function copyCode(btn) {
    var pre = btn.parentElement.querySelector("pre");
    var code = pre.textContent;
    navigator.clipboard.writeText(code).then(function() {
        btn.textContent = "Copied!";
        btn.classList.add("copied");
        setTimeout(function() {
            btn.textContent = "Copy";
            btn.classList.remove("copied");
        }, 2000);
    });
}

document.addEventListener("DOMContentLoaded", function() {
    document.querySelectorAll(".text-block pre").forEach(function(pre) {
        var wrapper = document.createElement("div");
        wrapper.className = "code-wrapper";
        pre.parentNode.insertBefore(wrapper, pre);
        wrapper.appendChild(pre);
        var btn = document.createElement("button");
        btn.className = "copy-btn";
        btn.textContent = "Copy";
        btn.onclick = function() { copyCode(btn); };
        wrapper.appendChild(btn);
    });
});
`
