// Package cli is the leaf of launch-room: command dispatch, the environment contract and exit codes.
// It imports nothing from launch-room, so every other package can import it without a cycle.
package cli

import (
	"context"
	"fmt"
	"io"
)

// Exit codes.
const (
	ExitOK    = 0
	ExitError = 1
	ExitUsage = 2
)

// Command is one `<prog> <name>` subcommand. → cmd/site/internal/sitecmd, cmd/form/internal/formcmd
type Command interface {
	Run(ctx context.Context, args []string) int
}

// Environment is everything the commands take from the OS. → internal/osenv
type Environment interface {
	Getenv(key string) string
	Stdout() io.Writer
	Stderr() io.Writer
}

// App dispatches `<prog> <command>` to the Command registered for it. main supplies the commands,
// so this package never imports them.
type App struct {
	prog     string
	usage    string
	env      Environment
	commands map[string]Command
}

func NewApp(prog, usage string, env Environment, commands map[string]Command) *App {
	return &App{
		prog:     prog,
		usage:    usage,
		env:      env,
		commands: commands,
	}
}

func (a *App) Run(ctx context.Context, args []string) int {
	if len(args) == 0 {
		fmt.Fprint(a.env.Stderr(), a.usage)
		return ExitUsage
	}

	switch args[0] {
	case "help", "-h", "--help":
		fmt.Fprint(a.env.Stdout(), a.usage)
		return ExitOK
	}

	cmd, ok := a.commands[args[0]]
	if !ok {
		fmt.Fprintf(a.env.Stderr(), "%s: unknown command %q\n%s", a.prog, args[0], a.usage)
		return ExitUsage
	}
	return cmd.Run(ctx, args[1:])
}

// Usage prints a subcommand's usage line and returns ExitUsage.
func Usage(env Environment, line string) int {
	fmt.Fprintln(env.Stderr(), "usage: "+line)
	return ExitUsage
}

// Fail prints err as "<prog>: <err>" and returns ExitError.
func Fail(env Environment, prog string, err error) int {
	fmt.Fprintf(env.Stderr(), "%s: %v\n", prog, err)
	return ExitError
}
