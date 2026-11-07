package videos

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mittalkanishk/launch-room/internal/media"
	"github.com/mittalkanishk/launch-room/internal/renderer"
	"github.com/mittalkanishk/launch-room/internal/run"
	"github.com/mittalkanishk/launch-room/internal/studio"
)

// The tests run the real studio service over HTTP with fakes for the renderer and ffmpeg, and git
// for real.

// fakeRenderer "renders" a scene as its source, and fails one that raises.
type fakeRenderer struct{}

func (fakeRenderer) Render(ctx context.Context, source string) (renderer.Result, error) {
	if strings.Contains(source, "raise") {
		return renderer.Result{}, errors.New("manim failed: scene.py line 6: ValueError")
	}
	return renderer.Result{Video: []byte(source), Subtitles: "1\n00:00:00,000 --> 00:00:01,000\nhello\n"}, nil
}

type fakePost struct {
	quiet  bool
	burned *bool // records whether the last join burned the subtitles in
}

func (p fakePost) JoinScenes(ctx context.Context, scenes, subtitles []string, burn bool, out string) error {
	if p.burned != nil {
		*p.burned = burn
	}
	if err := os.WriteFile(strings.TrimSuffix(out, ".mp4")+".srt", []byte("1\n00:00:00,000 --> 00:00:01,000\nhello\n"), 0o644); err != nil {
		return err
	}
	return os.WriteFile(out, []byte("joined"), 0o644)
}
func (fakePost) Finish(ctx context.Context, in, out, look string) error {
	return os.WriteFile(out, []byte("final"), 0o644)
}
func (p fakePost) Measure(ctx context.Context, in string) (media.Measurements, error) {
	m := media.Measurements{DurationSeconds: 95, Width: 1920, Height: 1080, HasAudio: true, LoudnessLUFS: -14, TruePeakDBTP: -1.5}
	if p.quiet {
		m.LoudnessLUFS = -24
	}
	return m, nil
}
func (fakePost) Keyframes(ctx context.Context, in, dir string, every int) ([]string, error) {
	f := filepath.Join(dir, "keyframe-01.jpg")
	return []string{f}, os.WriteFile(f, []byte("jpg"), 0o644)
}
func (fakePost) Sheet(ctx context.Context, in, subtitles, out string) error {
	return os.WriteFile(out, []byte("sheet"), 0o644)
}
func (fakePost) Motion(ctx context.Context, in, subtitles, out string) error {
	return os.WriteFile(out, []byte("motion"), 0o644)
}

func newTestWorkspace(t *testing.T, post fakePost) *Workspace {
	t.Helper()
	root := t.TempDir()
	svc, err := studio.NewService(filepath.Join(root, "studio"), map[string]studio.Renderer{"manim": fakeRenderer{}, "stage": fakeRenderer{}}, post)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(svc.Handler())
	t.Cleanup(server.Close)

	var out bytes.Buffer
	return NewWorkspace(Config{
		Root:        filepath.Join(root, "videos"),
		SiteSources: filepath.Join(root, "sites"),
		Public:      filepath.Join(root, "public"),
		Domain:      "example.test",
		Wait:        5 * time.Second,
		Poll:        20 * time.Millisecond,
	}, studio.NewClient(server.URL, http.DefaultClient), run.NewExec(&out, io.Discard))
}

// fillPlan writes a ready explainer of two scenes, as the producer and director would.
func fillPlan(t *testing.T, dir string) {
	t.Helper()
	write := func(name, content string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("video.json", `{"title":"An agent is a loop","platform":"youtube","length_seconds":120,"look":"neutral"}`)
	write("brief.md", "# Brief\n\n## Concept\nAgents are loops.\n\n## Audience\nDevelopers.\n\n## Hook\nYour agent never runs code.\n\n## Call to action\nRead steerloop.\n")
	write("script.md", "## 1. Title\nClaude Code is a loop.\n\n## 2. The code\nHere it is in Go.\n")
	os.Remove(filepath.Join(dir, "scenes", "01.js"))
	write("scenes/01-title.js", "import { scene } from \"stage\";\nexport default scene({ shots: [{ type: \"title\", title: \"A loop\", steps: [\"Claude Code is a loop.\"] }] });\n")
	write("scenes/02-code.py", "from launchroom import *\n\nclass Shot(Explainer):\n    def construct(self):\n        self.code(\"for {}\", steps=[Step(\"Here it is in Go.\", lines=(1,))])\n")
}

func TestTheWholeVideoWorkflow(t *testing.T) {
	w := newTestWorkspace(t, fakePost{})
	ctx := context.Background()

	dir, err := w.New(ctx, "loop")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "scenes", "01.js")); err != nil {
		t.Fatalf("the starter has no first scene: %v", err)
	}
	if check, _ := w.Check(ctx, "loop"); len(check.Problems) == 0 {
		t.Fatal("an empty starter plan passed the check")
	}
	if _, err := w.Render(ctx, "loop", nil); err == nil || !strings.Contains(err.Error(), "isn't ready") {
		t.Fatalf("render of an empty plan: err = %v", err)
	}

	fillPlan(t, dir)
	check, err := w.Check(ctx, "loop")
	if err != nil || len(check.Problems) != 0 || check.Scenes != 2 {
		t.Fatalf("check = %+v, %v", check, err)
	}

	results, err := w.Render(ctx, "loop", nil)
	if err != nil || len(results) != 2 || results[0].Status != "done" || results[1].File != "02-code.py" {
		t.Fatalf("render = %+v, %v", results, err)
	}
	for _, f := range []string{"shot-01.mp4", "shot-01.srt", "shot-01.jpg", "shot-02.jpg"} {
		if _, err := os.Stat(filepath.Join(dir, "renders", "scenes", f)); err != nil {
			t.Errorf("after the render, renders/scenes lacks %s (the director looks at the sheets there)", f)
		}
	}
	plan, _ := readPlan(dir)
	if plan.Scenes[0].Engine != "stage" || plan.Scenes[1].Engine != "manim" {
		t.Errorf("engines = %s, %s; want stage for .js and manim for .py", plan.Scenes[0].Engine, plan.Scenes[1].Engine)
	}
	strips, err := w.Motion(ctx, "loop", []int{2})
	if err != nil || len(strips) != 1 || strips[0].Status != "done" || filepath.Base(strips[0].File) != "shot-02-motion.jpg" {
		t.Fatalf("motion = %+v, %v", strips, err)
	}
	if _, err := os.Stat(filepath.Join(dir, "renders", "scenes", "shot-02-motion.jpg")); err != nil {
		t.Errorf("the motion strip wasn't fetched into renders/scenes: %v", err)
	}
	cut, err := w.Cut(ctx, "loop")
	if err != nil || !cut.Done || cut.PreviewURL != "https://preview-loop.example.test" {
		t.Fatalf("cut = %+v, %v", cut, err)
	}
	for _, step := range []func() (Outcome, error){
		func() (Outcome, error) { return w.Finish(ctx, "loop", "") },
		func() (Outcome, error) { return w.QC(ctx, "loop") },
	} {
		if out, err := step(); err != nil || !out.Done {
			t.Fatalf("step = %+v, %v", out, err)
		}
	}

	url, err := w.Publish(ctx, "loop")
	if err != nil || url != "https://loop.example.test" {
		t.Fatalf("publish = %q, %v", url, err)
	}
	if got, _ := os.ReadFile(filepath.Join(w.cfg.Public, "live", "loop", "video.mp4")); string(got) != "final" {
		t.Errorf("live video = %q, want the final render", got)
	}
}

func TestABrokenSceneReportsManimsError(t *testing.T) {
	w := newTestWorkspace(t, fakePost{})
	ctx := context.Background()
	dir, _ := w.New(ctx, "broken")
	fillPlan(t, dir)
	os.WriteFile(filepath.Join(dir, "scenes", "02-code.py"), []byte("from launchroom import *\n\nclass Shot(Explainer):\n    def construct(self):\n        self.title(\"x\", narration=\"y\")\n        raise ValueError\n"), 0o644)

	results, err := w.Render(ctx, "broken", []int{2})
	if err != nil || len(results) != 1 || results[0].Status != "failed" || !strings.Contains(results[0].Message, "line 6") {
		t.Fatalf("render of a broken scene = %+v, %v", results, err)
	}
	if _, err := w.Cut(ctx, "broken"); err == nil {
		t.Error("cut a video whose scenes aren't all rendered")
	}
}

func TestPublishNeedsAPassingQCOfTheFinal(t *testing.T) {
	w := newTestWorkspace(t, fakePost{quiet: true})
	ctx := context.Background()
	dir, _ := w.New(ctx, "quiet")
	fillPlan(t, dir)

	if _, err := w.Publish(ctx, "quiet"); err == nil {
		t.Error("published with no QC report")
	}
	w.Render(ctx, "quiet", nil)
	w.Cut(ctx, "quiet")
	w.Finish(ctx, "quiet", "")
	out, err := w.QC(ctx, "quiet")
	if err != nil {
		t.Fatal(err)
	}
	var report studio.Report
	content, _ := os.ReadFile(filepath.Join(dir, "renders", "qc", "report.json"))
	json.Unmarshal(content, &report)
	if report.Passed() || len(out.Files) == 0 {
		t.Fatalf("a -24 LUFS render passed QC: %+v", report)
	}
	if _, err := w.Publish(ctx, "quiet"); err == nil || !strings.Contains(err.Error(), "hasn't passed") {
		t.Errorf("publish after failed QC: err = %v", err)
	}
}

func TestSubtitlesSettingReachesTheCut(t *testing.T) {
	burned := true
	w := newTestWorkspace(t, fakePost{burned: &burned})
	ctx := context.Background()
	dir, _ := w.New(ctx, "clean")
	fillPlan(t, dir)
	if _, err := w.Render(ctx, "clean", nil); err != nil {
		t.Fatal(err)
	}

	// The starter burns them in; the editor turns that off in video.json; a typo is refused.
	for _, c := range []struct {
		subtitles string
		burn      bool
		problem   string
	}{
		{"burned", true, ""},
		{"", true, ""},
		{"none", false, ""},
		{"off", true, "unknown subtitles"},
	} {
		os.WriteFile(filepath.Join(dir, "video.json"), []byte(`{"title":"An agent is a loop","platform":"youtube","length_seconds":120,"look":"neutral","subtitles":"`+c.subtitles+`"}`), 0o644)
		burned = true
		_, err := w.Cut(ctx, "clean")
		switch {
		case c.problem != "":
			if err == nil || !strings.Contains(err.Error(), c.problem) {
				t.Errorf("subtitles %q: cut err = %v, want %q", c.subtitles, err, c.problem)
			}
		case err != nil:
			t.Errorf("subtitles %q: cut failed: %v", c.subtitles, err)
		case burned != c.burn:
			t.Errorf("subtitles %q: burned = %v, want %v", c.subtitles, burned, c.burn)
		}
	}
}

func TestNamesAreShared(t *testing.T) {
	w := newTestWorkspace(t, fakePost{})
	os.MkdirAll(filepath.Join(w.cfg.SiteSources, "loop"), 0o755)
	if _, err := w.New(context.Background(), "loop"); err == nil {
		t.Error("a video took a site's name")
	}
}
