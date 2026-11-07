package formcmd

import (
	"context"
	"fmt"

	"github.com/mittalkanishk/launch-room/internal/cli"
	"github.com/mittalkanishk/launch-room/internal/formspree"
)

// FormLister lists the sites' signup forms. → internal/formspree
type FormLister interface {
	Forms() ([]formspree.Form, error)
}

// List is `form list`: one line per site with its form's endpoint.
type List struct {
	env   cli.Environment
	forms FormLister
}

func NewList(env cli.Environment, forms FormLister) *List {
	return &List{
		env:   env,
		forms: forms,
	}
}

func (c *List) Run(ctx context.Context, args []string) int {
	if len(args) != 0 {
		return cli.Usage(c.env, "form list")
	}
	forms, err := c.forms.Forms()
	if err != nil {
		return cli.Fail(c.env, prog, err)
	}
	if len(forms) == 0 {
		fmt.Fprintln(c.env.Stdout(), "no forms yet")
		return cli.ExitOK
	}
	for _, f := range forms {
		fmt.Fprintf(c.env.Stdout(), "%s  %s\n", f.Site, f.Endpoint)
	}
	return cli.ExitOK
}
