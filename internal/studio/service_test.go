package studio

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mittalkanishk/launch-room/internal/media"
	"github.com/mittalkanishk/launch-room/internal/renderer"
)

// The tests run the real service over HTTP with fakes for the renderer and ffmpeg.

// fakeRenderer "renders" a scene as its source, and fails one that raises.
type fakeRenderer struct{}

func (fakeRenderer) Render(ctx context.Context, source string) (renderer.Result, error) {
	if strings.Contains(source, "raise") {
		return renderer.Result{}, errors.New("manim failed: scene.py line 3: ValueError")
	}
	return renderer.Result{Video: []byte(source), Subtitles: "1\n00:00:00,000 --> 00:00:01,000\nhello\n"}, nil
}

// fakePost joins and finishes by copying, and measures a passing render.
type fakePost struct{}

func (fakePost) JoinScenes(ctx context.Context, scenes, subtitles []string, burn bool, out string) error {
	var joined []string
	for _, s := range scenes {
		content, err := os.ReadFile(s)
		if err != nil {
			return err
		}
		joined = append(joined, string(content))
	}
	if !burn {
		joined = append(joined, "no subtitles")
	}
	if err := os.WriteFile(strings.TrimSuffix(out, ".mp4")+".srt", []byte("1\n00:00:00,000 --> 00:00:01,000\nhello\n"), 0o644); err != nil {
		return err
	}
	return os.WriteFile(out, []byte(strings.Join(joined, "|")), 0o644)
}
func (fakePost) Finish(ctx context.Context, in, out, look string) error { return copyFile(in, out) }
func (fakePost) Measure(ctx context.Context, in string) (media.Measurements, error) {
	return media.Measurements{DurationSeconds: 90, Width: 1920, Height: 1080, HasAudio: true, LoudnessLUFS: -14.2, TruePeakDBTP: -1.2}, nil
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

func copyFile(in, out string) error {
	data, err := os.ReadFile(in)
	if err != nil {
		return err
	}
	return os.WriteFile(out, data, 0o644)
}

func newTestStudio(t *testing.T) *Client {
	t.Helper()
	svc, err := NewService(t.TempDir(), map[string]Renderer{"manim": fakeRenderer{}, "stage": fakeRenderer{}}, fakePost{})
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(svc.Handler())
	t.Cleanup(server.Close)
	return NewClient(server.URL, http.DefaultClient)
}

// wait polls a job until it is done or failed.
func wait(t *testing.T, c *Client, job Job) Job {
	t.Helper()
	for range 200 {
		got, err := c.Job(context.Background(), job.ID)
		if err != nil {
			t.Fatal(err)
		}
		if got.Status == StatusDone || got.Status == StatusFailed {
			return got
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("job %s never finished", job.ID)
	return Job{}
}

func TestScenesCutFinishQC(t *testing.T) {
	c := newTestStudio(t)
	ctx := context.Background()

	for shot, source := range map[int]string{1: "scene one", 2: "scene two"} {
		done := wait(t, c, must(t)(c.Submit(ctx, Request{Kind: KindScene, Video: "loop", Shot: shot, Engine: []string{"", "manim", "stage"}[shot], Source: source})))
		if done.Status != StatusDone || len(done.Files) != 3 || done.Files[2] != "scenes/shot-0"+strconv.Itoa(shot)+".jpg" {
			t.Fatalf("scene %d = %+v", shot, done)
		}
	}
	if _, err := c.Submit(ctx, Request{Kind: KindFinish, Video: "loop", Look: "neutral"}); err == nil {
		t.Error("finishing before there is a cut was accepted")
	}

	cut := wait(t, c, must(t)(c.Submit(ctx, Request{Kind: KindCut, Video: "loop", Shots: []int{2, 1}})))
	if cut.Status != StatusDone || len(cut.Files) != 2 || cut.Files[0] != "cut.mp4" || cut.Files[1] != "cut.srt" {
		t.Fatalf("cut = %+v, want cut.mp4 and the joined cut.srt as its files", cut)
	}
	joined := filepath.Join(t.TempDir(), "cut.mp4")
	if err := c.Download(ctx, "loop", "cut.mp4", joined); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(joined); string(got) != "scene two|scene one" {
		t.Errorf("cut = %q, want the scenes in the order asked", got)
	}
	if cut.Started.IsZero() || cut.Started.Before(cut.Created) || cut.Updated.Before(cut.Started) {
		t.Errorf("a finished job's times are created %s, started %s, updated %s", cut.Created, cut.Started, cut.Updated)
	}

	clean := wait(t, c, must(t)(c.Submit(ctx, Request{Kind: KindCut, Video: "loop", Shots: []int{1, 2}, NoSubtitles: true})))
	if clean.Status != StatusDone {
		t.Fatalf("cut without subtitles = %+v", clean)
	}
	if err := c.Download(ctx, "loop", "cut.mp4", joined); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(joined); string(got) != "scene one|scene two|no subtitles" {
		t.Errorf("cut with no_subtitles = %q, want the join told not to burn them", got)
	}

	if finish := wait(t, c, must(t)(c.Submit(ctx, Request{Kind: KindFinish, Video: "loop", Look: "neutral"}))); finish.Status != StatusDone {
		t.Fatalf("finish = %+v", finish)
	}
	qc := wait(t, c, must(t)(c.Submit(ctx, Request{Kind: KindQC, Video: "loop"})))
	if qc.Status != StatusDone || len(qc.Files) != 2 {
		t.Fatalf("qc = %+v", qc)
	}
	report := filepath.Join(t.TempDir(), "report.json")
	if err := c.Download(ctx, "loop", "qc/report.json", report); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(report); !strings.Contains(string(got), `"render": "final.mp4"`) {
		t.Errorf("report = %s", got)
	}
}

func TestMotionStripNeedsARenderedScene(t *testing.T) {
	c := newTestStudio(t)
	ctx := context.Background()
	if _, err := c.Submit(ctx, Request{Kind: KindMotion, Video: "loop", Shot: 1}); err == nil {
		t.Error("a motion strip of an unrendered scene was accepted")
	}
	wait(t, c, must(t)(c.Submit(ctx, Request{Kind: KindScene, Video: "loop", Shot: 1, Engine: "stage", Source: "scene one"})))
	strip := wait(t, c, must(t)(c.Submit(ctx, Request{Kind: KindMotion, Video: "loop", Shot: 1})))
	if strip.Status != StatusDone || len(strip.Files) != 1 || strip.Files[0] != "scenes/shot-01-motion.jpg" {
		t.Fatalf("motion = %+v", strip)
	}
	if err := c.Download(ctx, "loop", strip.Files[0], filepath.Join(t.TempDir(), "m.jpg")); err != nil {
		t.Error(err)
	}
}

func TestAFailingSceneKeepsManimsError(t *testing.T) {
	c := newTestStudio(t)
	failed := wait(t, c, must(t)(c.Submit(context.Background(), Request{Kind: KindScene, Video: "loop", Shot: 1, Engine: "manim", Source: "raise"})))
	if failed.Status != StatusFailed || !strings.Contains(failed.Message, "line 3") {
		t.Errorf("failed scene = %+v, want Manim's error", failed)
	}
}

func TestRequestsAreChecked(t *testing.T) {
	c := newTestStudio(t)
	bad := []Request{
		{Kind: KindScene, Video: "../etc", Shot: 1, Engine: "manim", Source: "x"},
		{Kind: KindScene, Video: "loop", Shot: 0, Engine: "manim", Source: "x"},
		{Kind: KindScene, Video: "loop", Shot: 1, Engine: "three", Source: "x"},
		{Kind: KindCut, Video: "loop", Shots: []int{1}},
		{Kind: KindFinish, Video: "loop", Look: "sepia"},
		{Kind: "clip", Video: "loop"},
	}
	for _, req := range bad {
		if _, err := c.Submit(context.Background(), req); err == nil {
			t.Errorf("accepted %+v", req)
		}
	}
}

func TestChecksNeedAFullHDExplainerOfOneToTenMinutes(t *testing.T) {
	short := media.Measurements{DurationSeconds: 45, Width: 1920, Height: 1080, HasAudio: true, LoudnessLUFS: -14, TruePeakDBTP: -1.5}
	if (Report{Checks: Checks(short)}).Passed() {
		t.Error("a 45-second render passed")
	}
	vertical := short
	vertical.DurationSeconds, vertical.Width, vertical.Height = 90, 1080, 1920
	if (Report{Checks: Checks(vertical)}).Passed() {
		t.Error("a vertical render passed")
	}
}

func must(t *testing.T) func(Job, error) Job {
	return func(j Job, err error) Job {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
		return j
	}
}
