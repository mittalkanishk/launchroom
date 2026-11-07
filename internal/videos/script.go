package videos

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// A section of the script speaks for one scene. The producer writes 40 to 160 words each, and the
// voice says about 3 words a second (SpokenSeconds); the whole script should land near the
// length in video.json.
const (
	minSectionWords = 40
	maxSectionWords = 160
	wordsPerSecond  = 3
	lengthTolerance = 0.2 // the spoken length may miss the target by this much either way
)

// ScriptPreview renders script.md as a page at preview-<name>.<domain>/script/, so the room can
// read the narration, with each section's words and seconds, before the director writes scenes.
// It doesn't need the plan to be ready: a half-written script is what the room should see.
func (w *Workspace) ScriptPreview(ctx context.Context, name string) (string, error) {
	dir, err := w.Dir(name)
	if err != nil {
		return "", err
	}
	p, err := readPlan(dir)
	if err != nil {
		return "", err
	}
	if p.Script == "" {
		return "", fmt.Errorf("script.md has no narration yet")
	}
	page := scriptPage(name, p).html()

	target := filepath.Join(w.cfg.Public, "preview", name, "script")
	staging := target + ".new"
	os.RemoveAll(staging)
	if err := os.MkdirAll(staging, 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(staging, "index.html"), []byte(page), 0o644); err != nil {
		return "", err
	}
	if err := swap(staging, target); err != nil {
		return "", err
	}
	return w.url("preview-"+name) + "/script/", nil
}

// scriptPage lays the script out section by section, each with its word count and spoken seconds,
// and the whole against the target length.
func scriptPage(name string, p Plan) markdownPage {
	page := markdownPage{
		Title:    p.Meta.Title,
		Subtitle: fmt.Sprintf("Script · %s · target %ds", p.Meta.Platform, p.Meta.LengthSeconds),
	}
	if page.Title == "" {
		page.Title = name
	}
	spoken := p.SpokenSeconds()
	page.Summary = fmt.Sprintf("%d words, about %ds spoken", spokenWords(p.Script), spoken)
	if target := float64(p.Meta.LengthSeconds); target > 0 {
		diff := float64(spoken) - target
		page.Warn = diff > target*lengthTolerance || diff < -target*lengthTolerance
	}
	for _, s := range splitSections(p.Script) {
		words := spokenWords(s.body)
		section := pageSection{Body: render(s.heading + s.body)}
		if s.heading != "" {
			section.Note = fmt.Sprintf("%d words · about %ds", words, words/wordsPerSecond)
			section.Warn = words < minSectionWords || words > maxSectionWords
		}
		page.Sections = append(page.Sections, section)
	}
	return page
}

// spokenWords counts the words the voice will say: headings are not spoken, and Markdown marks on
// their own, like a table's pipes and rules, are not words.
func spokenWords(text string) int {
	n := 0
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "#") {
			continue
		}
		for _, f := range strings.Fields(line) {
			if strings.ContainsFunc(f, func(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }) {
				n++
			}
		}
	}
	return n
}

// section is a "## heading" and the text under it. Text before the first heading is a section
// with no heading, and no note.
type section struct {
	heading string // the heading line with its newline, empty before the first heading
	body    string
}

func splitSections(script string) []section {
	var sections []section
	current := section{}
	for _, line := range strings.SplitAfter(script, "\n") {
		if strings.HasPrefix(line, "## ") {
			if current.heading != "" || strings.TrimSpace(current.body) != "" {
				sections = append(sections, current)
			}
			current = section{heading: strings.TrimRight(line, "\n") + "\n"}
			continue
		}
		current.body += line
	}
	if current.heading != "" || strings.TrimSpace(current.body) != "" {
		sections = append(sections, current)
	}
	return sections
}
