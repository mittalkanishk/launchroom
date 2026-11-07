package videocmd

import (
	"context"
	"fmt"

	"github.com/mittalkanishk/launch-room/internal/cli"
	"github.com/mittalkanishk/launch-room/internal/prompts"
	"github.com/mittalkanishk/launch-room/internal/videos"
)

// SceneRenderer renders a video's scenes and reports how a pass is going. → internal/videos
type SceneRenderer interface {
	Render(ctx context.Context, name string, shots []int) ([]videos.SceneResult, error)
	Progress(ctx context.Context, name string) (videos.Progress, bool, error)
}

// Render is `video render <name> [--shots 1,3]`: the scenes rendered at full HD, narration
// included.
type Render struct {
	env      cli.Environment
	renderer SceneRenderer
}

func NewRender(env cli.Environment, renderer SceneRenderer) *Render {
	return &Render{env: env, renderer: renderer}
}

func (c *Render) Run(ctx context.Context, raw []string) int {
	a, ok := parseArgs(raw, true, false)
	if !ok {
		return cli.Usage(c.env, "video render <name> [--shots 1,3]")
	}
	results, err := c.renderer.Render(ctx, a.name, a.shots)
	if err != nil {
		return cli.Fail(c.env, prog, err)
	}
	out := c.env.Stdout()
	failed, running := 0, 0
	for _, r := range results {
		fmt.Fprintf(out, "scene %02d (%s): %s\n", r.Shot, r.File, r.Status)
		switch r.Status {
		case "failed":
			failed++
			fmt.Fprintf(out, "%s\n\n", r.Message)
		case "still rendering":
			running++
		}
	}
	if len(results) > running+failed {
		fmt.Fprintf(out, "look at every renders/scenes/shot-NN.jpg: a frame from each sentence of the scene\n")
	}
	if running > 0 {
		fmt.Fprintf(out, "%d still rendering; follow with: video status %s\n", running, a.name)
		if p, ok, err := c.renderer.Progress(ctx, a.name); err == nil && ok {
			printProgress(c.env, a.name, p)
		}
	}
	if failed > 0 {
		fmt.Fprintf(out, "%d failed: fix the scenes and render them again with --shots\n", failed)
		return cli.ExitError
	}
	return cli.ExitOK
}

// Scenes is `video scenes`: the scene library's reference, for writing a video's scenes.
type Scenes struct {
	env cli.Environment
}

func NewScenes(env cli.Environment) *Scenes { return &Scenes{env: env} }

func (c *Scenes) Run(ctx context.Context, raw []string) int {
	fmt.Fprint(c.env.Stdout(), prompts.Scenes)
	return cli.ExitOK
}

// takeFlag removes a yes/no flag from the arguments and reports whether it was there.
func takeFlag(raw []string, flag string) ([]string, bool) {
	var rest []string
	found := false
	for _, arg := range raw {
		if arg == flag {
			found = true
			continue
		}
		rest = append(rest, arg)
	}
	return rest, found
}
