package cli

import (
	"bytes"
	"context"
	"io"
	"slices"
	"testing"
)

type bufferEnv struct {
	stdout bytes.Buffer
	stderr bytes.Buffer
}

func (e *bufferEnv) Getenv(string) string { return "" }
func (e *bufferEnv) Stdout() io.Writer    { return &e.stdout }
func (e *bufferEnv) Stderr() io.Writer    { return &e.stderr }

type recordingCommand struct {
	args []string
}

func (c *recordingCommand) Run(ctx context.Context, args []string) int {
	c.args = args
	return ExitOK
}

func TestAppDispatch(t *testing.T) {
	env := &bufferEnv{}
	cmd := &recordingCommand{}
	app := NewApp("site", "usage text\n", env, map[string]Command{"preview": cmd})

	if code := app.Run(context.Background(), []string{"preview", "demo"}); code != ExitOK || !slices.Equal(cmd.args, []string{"demo"}) {
		t.Errorf("preview demo: code %d, args %v", code, cmd.args)
	}
	if code := app.Run(context.Background(), nil); code != ExitUsage {
		t.Errorf("no arguments: code %d, want ExitUsage", code)
	}
	if code := app.Run(context.Background(), []string{"nope"}); code != ExitUsage {
		t.Errorf("unknown command: code %d, want ExitUsage", code)
	}
	if code := app.Run(context.Background(), []string{"help"}); code != ExitOK || env.stdout.String() != "usage text\n" {
		t.Errorf("help: code %d, stdout %q", code, env.stdout.String())
	}
}
