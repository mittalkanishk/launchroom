// Command transcribe prints what is said in a voice note: `transcribe <audio file>`. The builder runs
// it on voice notes people send in the room; see internal/transcribe.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mittalkanishk/launch-room/internal/cli"
	"github.com/mittalkanishk/launch-room/internal/osenv"
	"github.com/mittalkanishk/launch-room/internal/settings"
	"github.com/mittalkanishk/launch-room/internal/transcribe"
)

// settingsFile is written by the controller's entrypoint; see internal/settings.
const settingsFile = "/home/node/.config/launch-room/settings.env"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := execute(ctx, osenv.Real{}, os.Args[1:])
	stop()
	os.Exit(code)
}

func execute(ctx context.Context, env cli.Environment, args []string) int {
	if len(args) != 1 || args[0] == "-h" || args[0] == "--help" {
		return cli.Usage(env, "transcribe <audio file>   print what is said in a voice note")
	}
	s, err := settings.Load(settingsFile, env)
	if err != nil {
		return cli.Fail(env, "transcribe", err)
	}

	client := transcribe.NewClient(transcribe.Config{
		APIKey:  s.OpenAIAPIKey,
		Model:   s.TranscribeModel,
		BaseURL: "https://api.openai.com/v1",
	}, &http.Client{Timeout: 2 * time.Minute})

	text, err := client.Transcribe(ctx, args[0])
	if err != nil {
		return cli.Fail(env, "transcribe", err)
	}
	fmt.Fprintln(env.Stdout(), text)
	return cli.ExitOK
}
