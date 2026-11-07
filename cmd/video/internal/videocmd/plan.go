package videocmd

import (
	"context"
	"fmt"

	"github.com/mittalkanishk/launch-room/internal/cli"
	"github.com/mittalkanishk/launch-room/internal/videos"
)

// Planner creates a video's folder and checks its plan. → internal/videos
type Planner interface {
	New(ctx context.Context, name string) (string, error)
	Check(ctx context.Context, name string) (videos.Checkup, error)
}

// New is `video new <name>`: the producer starts an explainer.
type New struct {
	env   cli.Environment
	plans Planner
}

func NewNew(env cli.Environment, plans Planner) *New {
	return &New{env: env, plans: plans}
}

func (c *New) Run(ctx context.Context, raw []string) int {
	a, ok := parseArgs(raw, false, false)
	if !ok {
		return cli.Usage(c.env, "video new <name>")
	}
	dir, err := c.plans.New(ctx, a.name)
	if err != nil {
		return cli.Fail(c.env, prog, err)
	}
	fmt.Fprintf(c.env.Stdout(), "created %s\nfill in brief.md, video.json and script.md (producer), then a scene per section in scenes/ (director; see: video scenes); check with: video check %s\n", dir, a.name)
	return cli.ExitOK
}

// Check is `video check <name>`: whether the plan is ready, and how long it will run.
type Check struct {
	env   cli.Environment
	plans Planner
}

func NewCheck(env cli.Environment, plans Planner) *Check {
	return &Check{env: env, plans: plans}
}

func (c *Check) Run(ctx context.Context, raw []string) int {
	a, ok := parseArgs(raw, false, false)
	if !ok {
		return cli.Usage(c.env, "video check <name>")
	}
	check, err := c.plans.Check(ctx, a.name)
	if err != nil {
		return cli.Fail(c.env, prog, err)
	}
	out := c.env.Stdout()
	fmt.Fprintf(out, "%d scenes, about %ds of narration, target %ds\n", check.Scenes, check.SpokenSeconds, check.TargetSeconds)
	if len(check.Problems) == 0 {
		fmt.Fprintln(out, "ready")
		return cli.ExitOK
	}
	fmt.Fprintln(out, "not ready:")
	for _, p := range check.Problems {
		fmt.Fprintf(out, "  - %s\n", p)
	}
	return cli.ExitError
}
