// Package formcmd is the `form` subcommands. Each one reads its arguments, asks the Formspree project
// (and, for add, the site workspace) to do the work, and prints the result.
package formcmd

import (
	"context"
	"fmt"

	"github.com/mittalkanishk/launch-room/internal/cli"
)

const prog = "form"

// endpointVar is the variable a site's code posts its signup form to; Vite reads it from .env.
const endpointVar = "VITE_SIGNUP_ENDPOINT"

// SiteFiles finds a site and sets values in its .env. → internal/sites
type SiteFiles interface {
	Dir(name string) (string, error)
	SetEnv(name, key, value string) error
}

// FormDeployer creates a site's signup form. → internal/formspree
type FormDeployer interface {
	Add(ctx context.Context, site string) (endpoint string, err error)
}

// Add is `form add <site>`: it creates the site's signup form and gives the site its endpoint.
type Add struct {
	env   cli.Environment
	sites SiteFiles
	forms FormDeployer
}

func NewAdd(env cli.Environment, sites SiteFiles, forms FormDeployer) *Add {
	return &Add{
		env:   env,
		sites: sites,
		forms: forms,
	}
}

func (c *Add) Run(ctx context.Context, args []string) int {
	if len(args) != 1 {
		return cli.Usage(c.env, "form add <site>")
	}
	site := args[0]

	// Check the site exists before creating a form for it.
	if _, err := c.sites.Dir(site); err != nil {
		return cli.Fail(c.env, prog, err)
	}
	endpoint, err := c.forms.Add(ctx, site)
	if err != nil {
		return cli.Fail(c.env, prog, err)
	}
	if err := c.sites.SetEnv(site, endpointVar, endpoint); err != nil {
		return cli.Fail(c.env, prog, err)
	}

	fmt.Fprintf(c.env.Stdout(), "signup form: %s\n", endpoint)
	fmt.Fprintf(c.env.Stdout(), "the site posts to import.meta.env.%s; run site preview %s to build it in\n", endpointVar, site)
	return cli.ExitOK
}
