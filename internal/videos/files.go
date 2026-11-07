package videos

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mittalkanishk/launch-room/internal/media"
)

// Meta is video.json: the title, platform and length are the producer's; the look and the
// subtitles are the editor's.
type Meta struct {
	Title         string `json:"title"`
	Platform      string `json:"platform"`
	LengthSeconds int    `json:"length_seconds"`
	Look          string `json:"look"`
	Subtitles     string `json:"subtitles"` // SubtitlesBurned (the default when empty) or SubtitlesNone
}

// The subtitles settings. The joined SRT is written beside the cut either way, for upload as
// captions; the setting decides whether the words are also drawn into the picture.
const (
	SubtitlesBurned = "burned"
	SubtitlesNone   = "none"
)

// BurnSubtitles reports whether the cut draws the subtitles into the picture.
func (m Meta) BurnSubtitles() bool { return m.Subtitles != SubtitlesNone }

// Plan is a video's files, read and checked together.
type Plan struct {
	Meta   Meta
	Brief  string
	Script string // the narration, without the starter's comments
	Scenes []Scene
}

// briefSections are the headings the producer must fill in.
var briefSections = []string{"Concept", "Audience", "Hook", "Call to action"}

var comments = regexp.MustCompile(`(?s)<!--.*?-->`)

// readPlan reads the video's files.
func readPlan(dir string) (Plan, error) {
	var p Plan
	if err := readJSON(filepath.Join(dir, "video.json"), &p.Meta); err != nil {
		return p, err
	}
	brief, err := os.ReadFile(filepath.Join(dir, "brief.md"))
	if err != nil {
		return p, err
	}
	p.Brief = string(brief)
	script, err := os.ReadFile(filepath.Join(dir, "script.md"))
	if err != nil {
		return p, err
	}
	p.Script = strings.TrimSpace(comments.ReplaceAllString(string(script), ""))
	p.Scenes, err = readScenes(dir)
	return p, err
}

// Problems lists what stops the plan going to rendering, for `video check` to report. An empty
// list means it is ready. Whether a scene's Python runs is found by rendering it.
func (p Plan) Problems() []string {
	var problems []string
	if p.Meta.Title == "" {
		problems = append(problems, "video.json: no title")
	}
	if _, ok := media.Looks[p.Meta.Look]; !ok {
		problems = append(problems, fmt.Sprintf("video.json: unknown look %q; the looks are %s", p.Meta.Look, strings.Join(media.LookNames(), ", ")))
	}
	switch p.Meta.Subtitles {
	case "", SubtitlesBurned, SubtitlesNone:
	default:
		problems = append(problems, fmt.Sprintf("video.json: unknown subtitles %q; write %q (drawn into the picture) or %q (only the SRT beside the cut)", p.Meta.Subtitles, SubtitlesBurned, SubtitlesNone))
	}
	for _, section := range briefSections {
		if !sectionFilled(p.Brief, section) {
			problems = append(problems, fmt.Sprintf("brief.md: the %q section is empty", section))
		}
	}
	if p.Script == "" {
		problems = append(problems, "script.md: no narration")
	}
	if len(p.Scenes) == 0 {
		problems = append(problems, "scenes/: no scenes; write scenes/01.js, 02.js...")
	}
	for i, s := range p.Scenes {
		if s.Shot != i+1 {
			problems = append(problems, fmt.Sprintf("scenes/%s: expected scene %02d here; number them 01, 02, 03... with no gaps or repeats", s.File, i+1))
		}
		problems = append(problems, sceneProblems(s)...)
	}
	return problems
}

// sceneProblems is what a scene's file is missing for its engine. Whether its code runs is found by
// rendering it.
func sceneProblems(s Scene) []string {
	var problems []string
	switch s.Engine {
	case "stage":
		if !strings.Contains(s.Source, "export default scene(") {
			problems = append(problems, fmt.Sprintf("scenes/%s: export default scene({ style, shots: [...] })", s.File))
		}
		if !strings.Contains(s.Source, "steps") {
			problems = append(problems, fmt.Sprintf("scenes/%s: nothing is narrated; every shot needs steps", s.File))
		}
	case "manim":
		if !strings.Contains(s.Source, "class Shot(") {
			problems = append(problems, fmt.Sprintf("scenes/%s: define class Shot(Explainer)", s.File))
		}
		if !strings.Contains(s.Source, "narrat") && !strings.Contains(s.Source, "Step(") {
			problems = append(problems, fmt.Sprintf("scenes/%s: nothing is narrated; use the ready-made shots or self.narrate", s.File))
		}
	}
	return problems
}

// SpokenSeconds estimates how long the script takes to say: the voice speaks about 3 words a second.
func (p Plan) SpokenSeconds() int {
	return spokenWords(p.Script) / wordsPerSecond
}

// sectionFilled reports whether a "## name" section of the brief has any text of its own.
func sectionFilled(brief, name string) bool {
	_, rest, ok := strings.Cut(brief, "## "+name)
	if !ok {
		return false
	}
	body, _, _ := strings.Cut(rest, "\n## ")
	for _, line := range strings.Split(body, "\n")[1:] {
		if t := strings.TrimSpace(line); t != "" && !isStarterHint(t) {
			return true
		}
	}
	return false
}

// isStarterHint is a line of guidance from the starter brief rather than the producer's own words.
func isStarterHint(line string) bool {
	hints := []string{"What happens in the first", "Who decided what", "One line per file"}
	for _, h := range hints {
		if strings.HasPrefix(line, h) {
			return true
		}
	}
	return false
}

func readJSON(path string, v any) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(content, v); err != nil {
		return fmt.Errorf("%s: %w", filepath.Base(path), err)
	}
	return nil
}
