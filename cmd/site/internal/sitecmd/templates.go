package sitecmd

import (
	"context"
	"fmt"

	"github.com/mittalkanishk/launch-room/internal/cli"
	"github.com/mittalkanishk/launch-room/internal/sites"
)

// TemplateLister lists the templates new sites can start from. → internal/sites
type TemplateLister interface {
	Templates() ([]sites.Template, error)
}

// Templates is `site templates`: one line per template with what it is for, so the builder can
// pick the one that fits the request.
type Templates struct {
	env   cli.Environment
	sites TemplateLister
}

func NewTemplates(env cli.Environment, sites TemplateLister) *Templates {
	return &Templates{
		env:   env,
		sites: sites,
	}
}

func (c *Templates) Run(ctx context.Context, args []string) int {
	if len(args) != 0 {
		return cli.Usage(c.env, "site templates")
	}
	templates, err := c.sites.Templates()
	if err != nil {
		return cli.Fail(c.env, prog, err)
	}
	for _, t := range templates {
		fmt.Fprintf(c.env.Stdout(), "%-12s %s\n", t.ID, t.Description)
	}
	return cli.ExitOK
}
