package videocmd

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/mittalkanishk/launch-room/internal/cli"
)

type bufferEnv struct {
	stdout bytes.Buffer
	stderr bytes.Buffer
}

func (e *bufferEnv) Getenv(string) string { return "" }
func (e *bufferEnv) Stdout() io.Writer    { return &e.stdout }
func (e *bufferEnv) Stderr() io.Writer    { return &e.stderr }

type fakeScriptPreviewer struct{ name string }

func (f *fakeScriptPreviewer) ScriptPreview(ctx context.Context, name string) (string, error) {
	f.name = name
	return "https://preview-" + name + ".example.test/script/", nil
}

func TestScriptPreviewTakesOneNameAfterPreview(t *testing.T) {
	ctx := context.Background()
	env := &bufferEnv{}
	preview := &fakeScriptPreviewer{}
	cmd := NewScript(env, preview)

	if code := cmd.Run(ctx, []string{"preview", "loop"}); code != cli.ExitOK || preview.name != "loop" {
		t.Fatalf("code = %d, previewed %q", code, preview.name)
	}
	if got := env.stdout.String(); got != "script: https://preview-loop.example.test/script/\n" {
		t.Errorf("printed %q", got)
	}
	for _, bad := range [][]string{nil, {"loop"}, {"preview"}, {"show", "loop"}, {"preview", "loop", "--look", "warm"}} {
		if code := cmd.Run(ctx, bad); code != cli.ExitUsage {
			t.Errorf("args %v: code = %d, want usage", bad, code)
		}
	}
}
