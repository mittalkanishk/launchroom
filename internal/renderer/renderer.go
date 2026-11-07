// Package renderer turns a scene's source into a video: the source in, an MP4 with its narration and
// the narration's subtitles out. There are two engines, each in its own keyless container apart
// from studio, because the source is code the agents wrote: Manim (Python, the launchroom library)
// and stage (JavaScript, anime.js in headless Chromium).
//
// Read in this order: renderer.go (rendering one scene), engines.go (Manim and stage), server.go
// (the HTTP API), client.go (studio's side).
package renderer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mittalkanishk/launch-room/internal/run"
)

// Result is one rendered scene.
type Result struct {
	Video     []byte `json:"video"`     // MP4, narration included
	Subtitles string `json:"subtitles"` // SRT of the narration
}

// Runner runs Manim. → internal/run
type Runner interface {
	Combined(ctx context.Context, c run.Cmd) (string, error)
}

// Engine is one way of turning a scene's source into a narrated video. → engines.go (Manim, Stage)
type Engine interface {
	SourceFile() string // the file name the source is written to
	Check(source string) error
	Command(dir string) run.Cmd // renders at media.Width x media.Height, media.FrameRate
	Outputs(dir string) (video, subtitles string, err error)
}

// Renderer renders scenes one at a time (each engine uses every core it gets).
type Renderer struct {
	workDir string
	engine  Engine
	runner  Runner
	timeout time.Duration
	slot    chan struct{}
}

func New(workDir string, engine Engine, runner Runner, timeout time.Duration) *Renderer {
	return &Renderer{
		workDir: workDir,
		engine:  engine,
		runner:  runner,
		timeout: timeout,
		slot:    make(chan struct{}, 1),
	}
}

// Render writes source to a scratch folder, renders it at full HD, and returns the video and
// subtitles. A failure carries the end of the engine's output, which names the line that broke.
func (r *Renderer) Render(ctx context.Context, source string) (Result, error) {
	if err := r.engine.Check(source); err != nil {
		return Result{}, err
	}

	r.slot <- struct{}{}
	defer func() { <-r.slot }()
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	dir, err := os.MkdirTemp(r.workDir, "scene-")
	if err != nil {
		return Result{}, err
	}
	defer os.RemoveAll(dir)
	if err := os.WriteFile(filepath.Join(dir, r.engine.SourceFile()), []byte(source), 0o644); err != nil {
		return Result{}, err
	}

	log, err := r.runner.Combined(ctx, r.engine.Command(dir))
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return Result{}, fmt.Errorf("the scene took longer than %s to render", r.timeout)
	}
	if err != nil {
		return Result{}, fmt.Errorf("the render failed:\n%s", tail(log, 1500))
	}

	videoFile, subtitleFile, err := r.engine.Outputs(dir)
	if err != nil {
		return Result{}, fmt.Errorf("%w:\n%s", err, tail(log, 800))
	}
	video, err := os.ReadFile(videoFile)
	if err != nil {
		return Result{}, err
	}
	subtitles, _ := os.ReadFile(subtitleFile)
	return Result{Video: video, Subtitles: string(subtitles)}, nil
}

func tail(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) > n {
		s = "..." + s[len(s)-n:]
	}
	return s
}
