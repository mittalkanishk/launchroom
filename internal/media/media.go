// Package media is the crew's post-production done by ffmpeg: joining an explainer's rendered scenes
// with their subtitles, applying one look and loudness to the cut (grading and mixing), and
// measuring the result for QC.
//
// Read in this order: media.go (the toolkit and its runner), then join.go, finish.go and measure.go.
package media

import (
	"context"

	"github.com/mittalkanishk/launch-room/internal/run"
)

// Runner runs ffmpeg and ffprobe. → internal/run
type Runner interface {
	Run(ctx context.Context, c run.Cmd) error
	Output(ctx context.Context, c run.Cmd) (string, error)
	Log(ctx context.Context, c run.Cmd) (string, error)
}

// Explainer frame: 16:9 full HD at 30 frames a second, for the scene renders and the cut alike.
const (
	Width     = 1920
	Height    = 1080
	FrameRate = 30
)

// Tools runs the ffmpeg steps.
type Tools struct {
	runner Runner
}

func NewTools(runner Runner) *Tools {
	return &Tools{runner: runner}
}

// ffmpeg is an ffmpeg command that overwrites its output and logs only problems and requested
// measurements.
func ffmpeg(args ...string) run.Cmd {
	return run.Cmd{Name: "ffmpeg", Args: append([]string{"-hide_banner", "-nostdin", "-y"}, args...)}
}
