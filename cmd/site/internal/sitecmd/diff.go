package sitecmd

import (
	"context"

	"github.com/mittalkanishk/launch-room/internal/cli"
)

// Differ writes what changed in a site since it last went live. → internal/sites
type Differ interface {
	Diff(ctx context.Context, name string) error
}

// Diff is `site diff <name>`.
type Diff struct {
	env   cli.Environment
	sites Differ
}

func NewDiff(env cli.Environment, sites Differ) *Diff {
	return &Diff{
		env:   env,
		sites: sites,
	}
}

func (c *Diff) Run(ctx context.Context, args []string) int {
	if len(args) != 1 {
		return cli.Usage(c.env, "site diff <name>")
	}
	if err := c.sites.Diff(ctx, args[0]); err != nil {
		return cli.Fail(c.env, prog, err)
	}
	return cli.ExitOK
}
