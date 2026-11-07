package videocmd

import (
	"context"
	"fmt"

	"github.com/mittalkanishk/launch-room/internal/cli"
)

// ScriptPreviewer puts the script on the video's preview host as a page. → internal/videos
type ScriptPreviewer interface {
	ScriptPreview(ctx context.Context, name string) (string, error)
}

// Script is `video script preview <name>`: the producer shows the room the narration, section by
// section with its length, before the director writes scenes.
type Script struct {
	env     cli.Environment
	preview ScriptPreviewer
}

func NewScript(env cli.Environment, preview ScriptPreviewer) *Script {
	return &Script{env: env, preview: preview}
}

func (c *Script) Run(ctx context.Context, raw []string) int {
	if len(raw) < 1 || raw[0] != "preview" {
		return cli.Usage(c.env, "video script preview <name>")
	}
	a, ok := parseArgs(raw[1:], false, false)
	if !ok {
		return cli.Usage(c.env, "video script preview <name>")
	}
	url, err := c.preview.ScriptPreview(ctx, a.name)
	if err != nil {
		return cli.Fail(c.env, prog, err)
	}
	fmt.Fprintf(c.env.Stdout(), "script: %s\n", url)
	return cli.ExitOK
}
