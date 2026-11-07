// Package videocmd is the `video` subcommands. Each reads its arguments, asks the video workspace
// to do the step, and prints what the crew posts in the room.
package videocmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mittalkanishk/launch-room/internal/cli"
	"github.com/mittalkanishk/launch-room/internal/videos"
)

const prog = "video"

// args is a subcommand's arguments: one video name and the options it allows.
type args struct {
	name  string
	shots []int  // --shots 1,3
	look  string // --look warm
}

// parseArgs reads "<name> [--shots 1,2] [--look x]" in any order, allowing only the options named.
func parseArgs(raw []string, allowShots, allowLook bool) (args, bool) {
	var a args
	var names []string
	for i := 0; i < len(raw); i++ {
		switch arg := raw[i]; {
		case arg == "--shots" && allowShots && i+1 < len(raw):
			shots, ok := parseShots(raw[i+1])
			if !ok {
				return a, false
			}
			a.shots = shots
			i++
		case arg == "--look" && allowLook && i+1 < len(raw):
			a.look = raw[i+1]
			i++
		case strings.HasPrefix(arg, "-"):
			return a, false
		default:
			names = append(names, arg)
		}
	}
	if len(names) != 1 {
		return a, false
	}
	a.name = names[0]
	return a, true
}

func parseShots(list string) ([]int, bool) {
	var shots []int
	for _, field := range strings.Split(list, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(field))
		if err != nil || n < 1 {
			return nil, false
		}
		shots = append(shots, n)
	}
	return shots, true
}

// printProgress reports a render pass that is still going: the line to post in the room, and
// when to post the next one.
func printProgress(env cli.Environment, name string, p videos.Progress) {
	fmt.Fprintln(env.Stdout(), p)
	fmt.Fprintf(env.Stdout(), "post that line in the room now; then every 2 minutes run once: video status %s  and post its first line, until nothing is running\n", name)
}

// printOutcome reports a render step: its files and preview, or that it is still running.
func printOutcome(env cli.Environment, name string, out videos.Outcome) {
	if !out.Done {
		progress := out.Job.Message
		if progress == "" {
			progress = out.Job.Status
		}
		fmt.Fprintf(env.Stdout(), "%s is still running (%s); check with: video status %s\n", out.Job.Kind, progress, name)
		return
	}
	for _, f := range out.Files {
		fmt.Fprintln(env.Stdout(), f)
	}
	if out.PreviewURL != "" {
		fmt.Fprintf(env.Stdout(), "preview: %s\n", out.PreviewURL)
	}
}
