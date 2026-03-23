package claude

import (
	"strings"
	"testing"
)

func TestRenderMarkdown_Heading(t *testing.T) {
	out := renderMarkdown("# Hello")
	if !strings.Contains(out, "<h1>Hello</h1>") {
		t.Errorf("expected <h1>, got: %s", out)
	}
}

func TestRenderMarkdown_Bold(t *testing.T) {
	out := renderMarkdown("**bold**")
	if !strings.Contains(out, "<strong>bold</strong>") {
		t.Errorf("expected <strong>, got: %s", out)
	}
}

func TestRenderMarkdown_CodeBlock(t *testing.T) {
	input := "```go\nfmt.Println(\"hi\")\n```"
	out := renderMarkdown(input)
	// chroma wraps code in a <pre> with class
	if !strings.Contains(out, "<pre") {
		t.Errorf("expected <pre> for code block, got: %s", out)
	}
	if !strings.Contains(out, "Println") {
		t.Errorf("expected code content preserved, got: %s", out)
	}
}

func TestRenderMarkdown_Empty(t *testing.T) {
	out := renderMarkdown("")
	if out != "" {
		t.Errorf("expected empty string, got: %s", out)
	}
}

func TestRenderMarkdown_PlainText(t *testing.T) {
	out := renderMarkdown("just plain text")
	if !strings.Contains(out, "just plain text") {
		t.Errorf("expected text preserved, got: %s", out)
	}
	// should be wrapped in <p>
	if !strings.Contains(out, "<p>") {
		t.Errorf("expected <p> wrapper, got: %s", out)
	}
}

func TestRenderMarkdown_HTMLEscaped(t *testing.T) {
	out := renderMarkdown("<script>alert('xss')</script>")
	if strings.Contains(out, "<script>") {
		t.Errorf("expected HTML to be escaped, got: %s", out)
	}
}

func TestRenderMarkdown_Table(t *testing.T) {
	input := "| A | B |\n|---|---|\n| 1 | 2 |"
	out := renderMarkdown(input)
	if !strings.Contains(out, "<table>") {
		t.Errorf("expected <table>, got: %s", out)
	}
}

func TestRenderMarkdown_InlineCode(t *testing.T) {
	out := renderMarkdown("use `fmt.Println`")
	if !strings.Contains(out, "<code>fmt.Println</code>") {
		t.Errorf("expected inline code, got: %s", out)
	}
}
