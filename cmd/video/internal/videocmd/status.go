package videocmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mittalkanishk/launch-room/internal/cli"
	"github.com/mittalkanishk/launch-room/internal/media"
	"github.com/mittalkanishk/launch-room/internal/studio"
	"github.com/mittalkanishk/launch-room/internal/videos"
)

// Tracker reports a video's studio jobs and fetches finished renders. → internal/videos
type Tracker interface {
	Status(ctx context.Context, name string) ([]studio.Job, []videos.Outcome, error)
}

// Status is `video status <name>`: the render pass that is going, if any, as one line for the
// room; then every studio job of the video, and the renders that finished since the last look.
type Status struct {
	env     cli.Environment
	tracker Tracker
}

func NewStatus(env cli.Environment, tracker Tracker) *Status {
	return &Status{env: env, tracker: tracker}
}

func (c *Status) Run(ctx context.Context, raw []string) int {
	a, ok := parseArgs(raw, false, false)
	if !ok {
		return cli.Usage(c.env, "video status <name>")
	}
	jobs, fetched, err := c.tracker.Status(ctx, a.name)
	if err != nil {
		return cli.Fail(c.env, prog, err)
	}
	out := c.env.Stdout()
	if p, ok := videos.RenderProgress(jobs, time.Now()); ok {
		printProgress(c.env, a.name, p)
	}
	if len(jobs) == 0 {
		fmt.Fprintln(out, "no studio jobs yet")
	}
	for _, j := range jobs {
		what := j.Kind
		if j.Shot > 0 {
			what = fmt.Sprintf("%s shot %d", j.Kind, j.Shot)
		}
		line := fmt.Sprintf("%-16s %-8s", what, j.Status)
		if j.Message != "" {
			line += "  " + j.Message
		}
		fmt.Fprintln(out, line)
	}
	for _, o := range fetched {
		printOutcome(c.env, a.name, o)
	}
	return cli.ExitOK
}

// Looks is `video looks`: the colour grades the editor can choose.
type Looks struct {
	env cli.Environment
}

func NewLooks(env cli.Environment) *Looks { return &Looks{env: env} }

func (c *Looks) Run(ctx context.Context, raw []string) int {
	fmt.Fprintln(c.env.Stdout(), strings.Join(media.LookNames(), "\n"))
	return cli.ExitOK
}
