// Package sitecmd is the `site` subcommands. Each one reads its arguments, asks the site workspace
// to do the work, and prints the result the agents post in the room.
package sitecmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/mittalkanishk/launch-room/internal/cli"
)

const prog = "site"

// Creator copies a template into a new site. → internal/sites
type Creator interface {
	Create(ctx context.Context, name, template string) (used string, err error)
}

// New is `site new <name> [--template <id>]`.
type New struct {
	env   cli.Environment
	sites Creator
}

func NewNew(env cli.Environment, sites Creator) *New {
	return &New{
		env:   env,
		sites: sites,
	}
}

func (c *New) Run(ctx context.Context, args []string) int {
	name, template, ok := parseNewArgs(args)
	if !ok {
		return cli.Usage(c.env, "site new <name> [--template <id>]")
	}
	used, err := c.sites.Create(ctx, name, template)
	if err != nil {
		return cli.Fail(c.env, prog, err)
	}
	fmt.Fprintf(c.env.Stdout(), "created site %s from template %s\n", name, used)
	return cli.ExitOK
}

// parseNewArgs reads one site name and an optional --template, in either order.
func parseNewArgs(args []string) (name, template string, ok bool) {
	var names []string
	for i := 0; i < len(args); i++ {
		switch arg := args[i]; {
		case arg == "--template" && i+1 < len(args):
			template = args[i+1]
			i++
		case strings.HasPrefix(arg, "--template="):
			template = strings.TrimPrefix(arg, "--template=")
		case strings.HasPrefix(arg, "-"):
			return "", "", false
		default:
			names = append(names, arg)
		}
	}
	if len(names) != 1 {
		return "", "", false
	}
	return names[0], template, true
}
