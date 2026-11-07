package videos

import (
	"bytes"
	"html"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

// The Markdown viewer: one of a video's Markdown files, rendered as a page the room can read. The
// script page is the first use (script.go); the brief and the shot plan can take the same layout.

// markdown renders Markdown as it was written, with tables and strikethrough, and no raw HTML.
var markdown = goldmark.New(goldmark.WithExtensions(extension.Table, extension.Strikethrough))

// render turns Markdown into HTML. Rendering can only fail on a broken writer, which a buffer
// never is.
func render(source string) string {
	var out bytes.Buffer
	markdown.Convert([]byte(source), &out)
	return out.String()
}

// markdownPage lays out a rendered file: the video's title, a line about the file, then the
// sections. Each section is an HTML fragment with its own note (word counts, timing) shown under
// its heading.
type markdownPage struct {
	Title    string
	Subtitle string
	Summary  string // one line about the whole file, e.g. the spoken length against the target
	Warn     bool   // the summary needs attention
	Sections []pageSection
}

type pageSection struct {
	Body string // rendered HTML
	Note string // shown under the heading, empty for none
	Warn bool   // the note needs attention
}

func (p markdownPage) html() string {
	var b strings.Builder
	b.WriteString(pageHead(p.Title))
	b.WriteString("<main>\n<header>\n<h1>" + html.EscapeString(p.Title) + "</h1>\n")
	if p.Subtitle != "" {
		b.WriteString("<p class=\"sub\">" + html.EscapeString(p.Subtitle) + "</p>\n")
	}
	if p.Summary != "" {
		b.WriteString("<p class=\"note" + warn(p.Warn) + "\">" + html.EscapeString(p.Summary) + "</p>\n")
	}
	b.WriteString("</header>\n")
	for _, s := range p.Sections {
		b.WriteString("<section>\n")
		if s.Note != "" {
			b.WriteString("<p class=\"note" + warn(s.Warn) + "\">" + html.EscapeString(s.Note) + "</p>\n")
		}
		b.WriteString(s.Body)
		b.WriteString("</section>\n")
	}
	b.WriteString("</main>\n</body>\n</html>\n")
	return b.String()
}

func warn(on bool) string {
	if on {
		return " warn"
	}
	return ""
}

func pageHead(title string) string {
	return `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>` + html.EscapeString(title) + `</title>
<style>
  :root { color-scheme: dark; }
  body { margin: 0; background: #0b0b0f; color: #e8e8ee; font-family: system-ui, sans-serif;
         font-size: 18px; line-height: 1.6; }
  main { max-width: 760px; margin: 0 auto; padding: 48px 20px 96px; }
  header { margin-bottom: 40px; }
  h1 { font-size: 34px; line-height: 1.2; margin: 0 0 8px; }
  h2 { font-size: 24px; margin: 40px 0 4px; }
  h3 { font-size: 20px; margin: 28px 0 4px; }
  .sub { margin: 0; color: #9a9aa8; }
  .note { margin: 0 0 12px; font-size: 14px; color: #8fd3a8; }
  .note.warn { color: #ff8e8e; }
  section { border-top: 1px solid #22222c; }
  section > h2:first-of-type { margin-top: 32px; }
  p { margin: 0 0 16px; }
  a { color: #c7c7ff; }
  code { font-family: ui-monospace, monospace; font-size: 0.9em; background: #16161e; padding: 2px 6px; border-radius: 4px; }
  pre { background: #16161e; padding: 16px; border-radius: 8px; overflow-x: auto; }
  pre code { background: none; padding: 0; }
  blockquote { margin: 0 0 16px; padding-left: 16px; border-left: 3px solid #44445a; color: #c0c0cc; }
  table { border-collapse: collapse; margin: 0 0 16px; }
  td, th { border: 1px solid #2a2a36; padding: 6px 12px; text-align: left; }
  img { max-width: 100%; }
</style>
</head>
<body>
`
}
