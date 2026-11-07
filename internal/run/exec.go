// Package run runs the external tools the launch-room commands drive: git, the package managers and
// the Formspree CLI.
package run

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Cmd is one external command.
type Cmd struct {
	Dir  string   // working directory; empty means the current one
	Name string   // program
	Args []string // arguments
	Env  []string // KEY=VALUE pairs added to this process's environment
}

// String is the command line, for error messages. It leaves out Env, which can hold secrets.
func (c Cmd) String() string {
	return strings.Join(append([]string{c.Name}, c.Args...), " ")
}

// Exec runs commands for real: Run streams their output, Output captures their standard output.
// Standard error always goes to stderr.
type Exec struct {
	stdout io.Writer
	stderr io.Writer
}

func NewExec(stdout, stderr io.Writer) *Exec {
	return &Exec{
		stdout: stdout,
		stderr: stderr,
	}
}

func (e *Exec) Run(ctx context.Context, c Cmd) error {
	cmd := e.command(ctx, c)
	cmd.Stdout = e.stdout
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s: %w", c, err)
	}
	return nil
}

func (e *Exec) Output(ctx context.Context, c Cmd) (string, error) {
	out, err := e.command(ctx, c).Output()
	if err != nil {
		return "", fmt.Errorf("%s: %w", c, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// Log runs c and returns what it wrote to standard error, where ffmpeg writes its measurements.
// On failure the error carries the end of that log.
func (e *Exec) Log(ctx context.Context, c Cmd) (string, error) {
	cmd := e.command(ctx, c)
	var log strings.Builder
	cmd.Stderr = &log
	if err := cmd.Run(); err != nil {
		tail := log.String()
		if len(tail) > 600 {
			tail = tail[len(tail)-600:]
		}
		return "", fmt.Errorf("%s: %w: %s", c, err, strings.TrimSpace(tail))
	}
	return log.String(), nil
}

// Combined runs c and returns its standard output and error together, as a person would see them,
// for tools like Manim that report errors on either.
func (e *Exec) Combined(ctx context.Context, c Cmd) (string, error) {
	cmd := e.command(ctx, c)
	var out strings.Builder
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		return out.String(), fmt.Errorf("%s: %w", c, err)
	}
	return out.String(), nil
}

// command is c as an exec.Cmd, killed if ctx ends.
func (e *Exec) command(ctx context.Context, c Cmd) *exec.Cmd {
	cmd := exec.CommandContext(ctx, c.Name, c.Args...)
	cmd.Dir = c.Dir
	cmd.Env = append(os.Environ(), c.Env...)
	cmd.Stderr = e.stderr
	return cmd
}
