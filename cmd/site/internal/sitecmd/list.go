package sitecmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/mittalkanishk/launch-room/internal/cli"
	"github.com/mittalkanishk/launch-room/internal/sites"
)

// Lister lists the sites. → internal/sites
type Lister interface {
	List(ctx context.Context) ([]sites.Entry, error)
}

// List is `site list`: one line per site with its preview and live URLs.
type List struct {
	env   cli.Environment
	sites Lister
}

func NewList(env cli.Environment, sites Lister) *List {
	return &List{
		env:   env,
		sites: sites,
	}
}

func (c *List) Run(ctx context.Context, args []string) int {
	if len(args) != 0 {
		return cli.Usage(c.env, "site list")
	}
	entries, err := c.sites.List(ctx)
	if err != nil {
		return cli.Fail(c.env, prog, err)
	}
	if len(entries) == 0 {
		fmt.Fprintln(c.env.Stdout(), "no sites yet")
		return cli.ExitOK
	}
	for _, e := range entries {
		fmt.Fprintln(c.env.Stdout(), line(e))
	}
	return cli.ExitOK
}

func line(e sites.Entry) string {
	parts := []string{e.Name}
	if e.PreviewURL != "" {
		parts = append(parts, "preview: "+e.PreviewURL)
	}
	if e.LiveURL != "" {
		parts = append(parts, "live: "+e.LiveURL)
	}
	if e.Unreleased > 0 {
		parts = append(parts, fmt.Sprintf("(%d preview(s) not live)", e.Unreleased))
	}
	return strings.Join(parts, "  ")
}
