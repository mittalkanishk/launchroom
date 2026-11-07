// Command roles gives the launch-room agents their roles: `roles install` writes the builder prompt
// for Claude Code and the reviewer prompt for Codex into the home folder ($HOME).
//
// Read in this order:
//
//	main.go                       which concrete type fills each need (newApp)
//	internal/rolescmd             the install subcommand
//	../../internal/roles          which prompt goes to which agent, and where
//	../../internal/prompts        the prompts themselves (builder.md, reviewer.md)
package main

import (
	"context"
	"os"

	"github.com/mittalkanishk/launch-room/cmd/roles/internal/rolescmd"
	"github.com/mittalkanishk/launch-room/internal/cli"
	"github.com/mittalkanishk/launch-room/internal/osenv"
	"github.com/mittalkanishk/launch-room/internal/roles"
)

const usage = `usage: roles install   write the builder and reviewer prompts into Claude Code's and Codex's homes
`

func main() {
	env := osenv.Real{}
	os.Exit(newApp(env).Run(context.Background(), os.Args[1:]))
}

// newApp is the one place the concrete types are chosen: the installer writes into $HOME.
func newApp(env cli.Environment) *cli.App {
	return cli.NewApp("roles", usage, env, map[string]cli.Command{
		"install": rolescmd.NewInstall(env, roles.NewInstaller(env.Getenv("HOME"))),
	})
}
