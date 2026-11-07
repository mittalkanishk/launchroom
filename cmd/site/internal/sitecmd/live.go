package sitecmd

import (
	"context"

	"github.com/mittalkanishk/launch-room/internal/cli"
	"github.com/mittalkanishk/launch-room/internal/sites"
)

// Launcher puts a site's current preview live. → internal/sites
type Launcher interface {
	Live(ctx context.Context, name string) (sites.Published, error)
}

// Live is `site live <name>`. The approval gate (cmd/gate) makes Claude Code ask the room before it
// may run this.
type Live struct {
	env   cli.Environment
	sites Launcher
}

func NewLive(env cli.Environment, sites Launcher) *Live {
	return &Live{
		env:   env,
		sites: sites,
	}
}

func (c *Live) Run(ctx context.Context, args []string) int {
	if len(args) != 1 {
		return cli.Usage(c.env, "site live <name>")
	}
	published, err := c.sites.Live(ctx, args[0])
	if err != nil {
		return cli.Fail(c.env, prog, err)
	}
	printPublished(c.env, "live", published)
	return cli.ExitOK
}
