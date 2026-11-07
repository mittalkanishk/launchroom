// Package rolescmd is the `roles` subcommands.
package rolescmd

import (
	"context"
	"fmt"

	"github.com/mittalkanishk/launch-room/internal/cli"
)

const prog = "roles"

// RoleWriter writes the agents' role prompts into a home folder. → internal/roles
type RoleWriter interface {
	Install() (written []string, err error)
}

// Install is `roles install`, which the controller's entrypoint runs on every start.
type Install struct {
	env   cli.Environment
	roles RoleWriter
}

func NewInstall(env cli.Environment, roles RoleWriter) *Install {
	return &Install{
		env:   env,
		roles: roles,
	}
}

func (c *Install) Run(ctx context.Context, args []string) int {
	if len(args) != 0 {
		return cli.Usage(c.env, "roles install")
	}
	written, err := c.roles.Install()
	if err != nil {
		return cli.Fail(c.env, prog, err)
	}
	for _, path := range written {
		fmt.Fprintf(c.env.Stdout(), "wrote %s\n", path)
	}
	return cli.ExitOK
}
