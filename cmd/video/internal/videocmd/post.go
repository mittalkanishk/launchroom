package videocmd

import (
	"context"

	"github.com/mittalkanishk/launch-room/internal/cli"
	"github.com/mittalkanishk/launch-room/internal/videos"
)

// Editing is the editor's steps: cutting, finishing and measuring. → internal/videos
type Editing interface {
	Cut(ctx context.Context, name string) (videos.Outcome, error)
	Finish(ctx context.Context, name, look string) (videos.Outcome, error)
	QC(ctx context.Context, name string) (videos.Outcome, error)
}

// Cut is `video cut <name>`: the scenes joined, each as last rendered, with subtitles burned in.
type Cut struct {
	env  cli.Environment
	edit Editing
}

func NewCut(env cli.Environment, edit Editing) *Cut { return &Cut{env: env, edit: edit} }

func (c *Cut) Run(ctx context.Context, raw []string) int {
	a, ok := parseArgs(raw, false, false)
	if !ok {
		return cli.Usage(c.env, "video cut <name>")
	}
	out, err := c.edit.Cut(ctx, a.name)
	return report(c.env, a.name, out, err)
}

// Finish is `video finish <name> [--look warm]`: one look over the whole cut, and social loudness.
type Finish struct {
	env  cli.Environment
	edit Editing
}

func NewFinish(env cli.Environment, edit Editing) *Finish { return &Finish{env: env, edit: edit} }

func (c *Finish) Run(ctx context.Context, raw []string) int {
	a, ok := parseArgs(raw, false, true)
	if !ok {
		return cli.Usage(c.env, "video finish <name> [--look <look>]")
	}
	out, err := c.edit.Finish(ctx, a.name, a.look)
	return report(c.env, a.name, out, err)
}

// QC is `video qc <name>`: the latest render measured against the rules, with keyframes.
type QC struct {
	env  cli.Environment
	edit Editing
}

func NewQC(env cli.Environment, edit Editing) *QC { return &QC{env: env, edit: edit} }

func (c *QC) Run(ctx context.Context, raw []string) int {
	a, ok := parseArgs(raw, false, false)
	if !ok {
		return cli.Usage(c.env, "video qc <name>")
	}
	out, err := c.edit.QC(ctx, a.name)
	return report(c.env, a.name, out, err)
}

// report prints a step's outcome and turns it into an exit code.
func report(env cli.Environment, name string, out videos.Outcome, err error) int {
	if err != nil {
		return cli.Fail(env, prog, err)
	}
	printOutcome(env, name, out)
	return cli.ExitOK
}
