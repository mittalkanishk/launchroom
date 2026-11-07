package renderer

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mittalkanishk/launch-room/internal/run"
)

// fakeRunner stands in for manim and node: it writes the files the engine would, or fails like a
// broken scene does.
type fakeRunner struct {
	fail bool
	cmds []run.Cmd
}

func (f *fakeRunner) Combined(ctx context.Context, c run.Cmd) (string, error) {
	f.cmds = append(f.cmds, c)
	if f.fail {
		return "Traceback\n  File \"scene.py\", line 7\nNameError: name 'Foo' is not defined", errors.New("exit status 1")
	}
	var video string
	switch c.Name {
	case "manim":
		video = filepath.Join(c.Dir, "media", "videos", "scene", "1080p30", "Shot.mp4")
	case "node":
		video = filepath.Join(c.Dir, "out", "scene.mp4")
	}
	os.MkdirAll(filepath.Dir(video), 0o755)
	os.WriteFile(video, []byte("mp4"), 0o644)
	os.WriteFile(strings.TrimSuffix(video, ".mp4")+".srt", []byte("1\n00:00:00,000 --> 00:00:01,000\nhi\n"), 0o644)
	return "", nil
}

func TestBothEnginesRenderThroughTheAPI(t *testing.T) {
	cases := []struct {
		engine Engine
		source string
		tool   string
	}{
		{Manim{}, "class Shot(Explainer):\n    pass\n", "manim"},
		{Stage{Script: "/opt/stage/runtime/render.mjs"}, "export default scene({ shots: [] });", "node"},
	}
	for _, c := range cases {
		runner := &fakeRunner{}
		server := httptest.NewServer(New(t.TempDir(), c.engine, runner, time.Minute).Handler())
		result, err := NewClient(server.URL, http.DefaultClient).Render(context.Background(), c.source)
		server.Close()
		if err != nil || string(result.Video) != "mp4" || !strings.Contains(result.Subtitles, "hi") {
			t.Fatalf("%s: render = %+v, %v", c.tool, result, err)
		}
		if got := runner.cmds[0]; got.Name != c.tool || !strings.Contains(strings.Join(got.Args, " "), "1920") {
			t.Errorf("%s: ran %s %v, want full HD", c.tool, got.Name, got.Args)
		}
	}
}

func TestABrokenSceneReturnsTheEnginesError(t *testing.T) {
	r := New(t.TempDir(), Manim{}, &fakeRunner{fail: true}, time.Minute)
	_, err := r.Render(context.Background(), "class Shot(Explainer):\n    Foo()\n")
	if err == nil || !strings.Contains(err.Error(), "line 7") {
		t.Errorf("err = %v, want the line that broke", err)
	}
}

func TestSourcesAreCheckedForTheirEngine(t *testing.T) {
	if _, err := New(t.TempDir(), Stage{}, &fakeRunner{}, time.Minute).Render(context.Background(), "class Shot: pass"); err == nil {
		t.Error("stage accepted a Python scene")
	}
	if _, err := New(t.TempDir(), Manim{}, &fakeRunner{}, time.Minute).Render(context.Background(), "export default scene({})"); err == nil {
		t.Error("manim accepted a JavaScript scene")
	}
}
