package claude

import (
	"bytes"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
)

var md = goldmark.New(
	goldmark.WithExtensions(
		extension.Table,
		extension.Strikethrough,
		extension.Linkify,
		highlighting.NewHighlighting(
			highlighting.WithStyle("monokai"),
			highlighting.WithFormatOptions(
				chromahtml.WithClasses(false),
				chromahtml.WithLineNumbers(false),
			),
		),
	),
)

func renderMarkdown(text string) string {
	if text == "" {
		return ""
	}
	var buf bytes.Buffer
	if err := md.Convert([]byte(text), &buf); err != nil {
		return text
	}
	return buf.String()
}
