package sitecmd

import (
	"context"
	"fmt"

	"github.com/mittalkanishk/launch-room/internal/cli"
	"github.com/mittalkanishk/launch-room/internal/sites"
)

// Previewer builds a site and publishes it at its preview URL. → internal/sites
type Previewer interface {
	Preview(ctx context.Context, name string) (sites.Published, error)
}

// Preview is `site preview <name>`.
type Preview struct {
	env   cli.Environment
	sites Previewer
}

func NewPreview(env cli.Environment, sites Previewer) *Preview {
	return &Preview{
		env:   env,
		sites: sites,
	}
}

func (c *Preview) Run(ctx context.Context, args []string) int {
	if len(args) != 1 {
		return cli.Usage(c.env, "site preview <name>")
	}
	published, err := c.sites.Preview(ctx, args[0])
	if err != nil {
		return cli.Fail(c.env, prog, err)
	}
	printPublished(c.env, "preview", published)
	return cli.ExitOK
}

// printPublished prints where the site is, and its editable deck if it has one.
func printPublished(env cli.Environment, stage string, p sites.Published) {
	fmt.Fprintf(env.Stdout(), "%s: %s\n", stage, p.URL)
	if p.DeckURL != "" {
		fmt.Fprintf(env.Stdout(), "deck: %s\n", p.DeckURL)
	}
}
