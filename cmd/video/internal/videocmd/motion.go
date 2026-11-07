package videocmd

import (
	"context"
	"fmt"

	"github.com/mittalkanishk/launch-room/internal/cli"
	"github.com/mittalkanishk/launch-room/internal/videos"
)

// MotionMaker makes a scene's motion strip. → internal/videos
type MotionMaker interface {
	Motion(ctx context.Context, name string, shots []int) ([]videos.SceneResult, error)
}

// Motion is `video motion <name> [--shots 1,3]`: for each rendered scene, one picture with a row
// of frames per sentence, left to right in time, so a reviewer can judge movement rather than the
// sheet's one frame per sentence.
type Motion struct {
	env   cli.Environment
	maker MotionMaker
}

func NewMotion(env cli.Environment, maker MotionMaker) *Motion {
	return &Motion{env: env, maker: maker}
}

func (c *Motion) Run(ctx context.Context, raw []string) int {
	a, ok := parseArgs(raw, true, false)
	if !ok {
		return cli.Usage(c.env, "video motion <name> [--shots 1,3]")
	}
	results, err := c.maker.Motion(ctx, a.name, a.shots)
	if err != nil {
		return cli.Fail(c.env, prog, err)
	}
	out := c.env.Stdout()
	failed, running := 0, 0
	for _, r := range results {
		switch r.Status {
		case "failed":
			failed++
			fmt.Fprintf(out, "scene %02d: failed: %s\n", r.Shot, r.Message)
		case "still rendering":
			running++
			fmt.Fprintf(out, "scene %02d: still rendering; follow with: video status %s\n", r.Shot, a.name)
		default:
			fmt.Fprintf(out, "scene %02d: %s\n", r.Shot, r.File)
		}
	}
	if len(results) > running+failed {
		fmt.Fprintln(out, "open each strip as an image: one row per sentence, eight frames left to right across the time it is said; judge the path, not the endpoints")
	}
	if failed > 0 {
		return cli.ExitError
	}
	return cli.ExitOK
}
