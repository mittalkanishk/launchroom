// Command video is the video crew's workflow for explainers: the producer plans, the director writes
// and renders one Manim scene per section, the editor cuts, finishes and checks, and a person in the
// room approves the publish. The rendering and post-production run in studio; this command plans
// the work and brings the results back.
//
// Read in this order:
//
//	main.go                       which concrete type fills each need (newApp)
//	internal/videocmd             one file per group of subcommands
//	../../internal/videos         the plan, the steps, previews and publishing
//	../../internal/studio         the service that does the work (client.go is this side of it)
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mittalkanishk/launch-room/cmd/video/internal/videocmd"
	"github.com/mittalkanishk/launch-room/internal/cli"
	"github.com/mittalkanishk/launch-room/internal/osenv"
	"github.com/mittalkanishk/launch-room/internal/run"
	"github.com/mittalkanishk/launch-room/internal/settings"
	"github.com/mittalkanishk/launch-room/internal/studio"
	"github.com/mittalkanishk/launch-room/internal/videos"
)

// settingsFile is written by the controller's entrypoint; see internal/settings.
const settingsFile = "/home/node/.config/launch-room/settings.env"

const usage = `usage: video new <name>                              start an explainer (producer)
       video script preview <name>                   the script as a page for the room to read
       video check <name>                            is the plan ready, and how long
       video scenes                                  the scene library, for writing scenes (director)
       video render <name> [--shots 1,3]             the scenes, narrated, full HD
       video motion <name> [--shots 1,3]             a frame strip per sentence, to judge movement (reviewer)
       video status <name>                           studio jobs, and renders that finished
       video cut <name>                              the scenes joined; subtitles burned in unless video.json says "subtitles": "none" (editor)
       video finish <name> [--look <look>]           one look and platform loudness
       video qc <name>                               measure the latest render, with keyframes
       video publish <name>                          put the QC-passed final live; needs approval
       video looks                                   the looks the editor can choose
`

func main() {
	// Stopping video also stops its wait for studio; studio's jobs carry on.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := execute(ctx, osenv.Real{}, os.Args[1:])
	stop()
	os.Exit(code)
}

func execute(ctx context.Context, env cli.Environment, args []string) int {
	s, err := settings.Load(settingsFile, env)
	if err != nil {
		return cli.Fail(env, "video", err)
	}
	return newApp(env, s).Run(ctx, args)
}

// newApp is the one place the concrete types are chosen: studio over HTTP, and git for real. Steps
// wait up to 100 seconds for studio, inside the agents' two-minute command limit.
func newApp(env cli.Environment, s settings.Settings) *cli.App {
	client := studio.NewClient(s.StudioURL, &http.Client{Timeout: 3 * time.Minute})
	workspace := videos.NewWorkspace(videos.Config{
		Root:        s.VideosDir,
		SiteSources: s.SitesSource,
		Public:      s.SitesPublic,
		Domain:      s.SitesDomain,
		Wait:        100 * time.Second,
		Poll:        2 * time.Second,
	}, client, run.NewExec(env.Stdout(), env.Stderr()))

	return cli.NewApp("video", usage, env, map[string]cli.Command{
		"new":     videocmd.NewNew(env, workspace),
		"script":  videocmd.NewScript(env, workspace),
		"check":   videocmd.NewCheck(env, workspace),
		"render":  videocmd.NewRender(env, workspace),
		"scenes":  videocmd.NewScenes(env),
		"motion":  videocmd.NewMotion(env, workspace),
		"status":  videocmd.NewStatus(env, workspace),
		"cut":     videocmd.NewCut(env, workspace),
		"finish":  videocmd.NewFinish(env, workspace),
		"qc":      videocmd.NewQC(env, workspace),
		"publish": videocmd.NewPublish(env, workspace),
		"looks":   videocmd.NewLooks(env),
	})
}
