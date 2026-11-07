package videocmd

import (
	"context"
	"fmt"

	"github.com/mittalkanishk/launch-room/internal/cli"
)

// Publisher puts a QC-passed final live. → internal/videos
type Publisher interface {
	Publish(ctx context.Context, name string) (string, error)
}

// Publish is `video publish <name>`. The approval gate asks the room before an agent may run it.
type Publish struct {
	env       cli.Environment
	publisher Publisher
}

func NewPublish(env cli.Environment, publisher Publisher) *Publish {
	return &Publish{env: env, publisher: publisher}
}

func (c *Publish) Run(ctx context.Context, raw []string) int {
	a, ok := parseArgs(raw, false, false)
	if !ok {
		return cli.Usage(c.env, "video publish <name>")
	}
	url, err := c.publisher.Publish(ctx, a.name)
	if err != nil {
		return cli.Fail(c.env, prog, err)
	}
	fmt.Fprintf(c.env.Stdout(), "live: %s\n", url)
	return cli.ExitOK
}
