// Command site is the launch-room agents' site workflow: copy the template into a new site, build and
// publish previews, show what changed, and put approved previews live.
//
// Read in this order:
//
//	main.go                       which concrete type fills each need (newApp)
//	internal/sitecmd              one file per subcommand: arguments in, result out
//	../../internal/sites          what each subcommand does to a site
//	../../internal/run            how git and the package managers are run
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/mittalkanishk/launch-room/cmd/site/internal/sitecmd"
	"github.com/mittalkanishk/launch-room/internal/cli"
	"github.com/mittalkanishk/launch-room/internal/osenv"
	"github.com/mittalkanishk/launch-room/internal/run"
	"github.com/mittalkanishk/launch-room/internal/settings"
	"github.com/mittalkanishk/launch-room/internal/sites"
)

// settingsFile is written by the controller's entrypoint; see internal/settings.
const settingsFile = "/home/node/.config/launch-room/settings.env"

// defaultTemplate is the template `site new` uses when none is chosen.
const defaultTemplate = "react-vite"

const usage = `usage: site templates                      list the templates and what each is for
       site new <name> [--template <id>]   copy a template (default react-vite) into a new site
       site preview <name>                 build the site and publish it at its preview URL
       site diff <name>                    show what changed since the site last went live (or since the template)
       site live <name>                    put the current preview live; only after a person in the room approves
       site list                           list the sites with their preview and live URLs
`

func main() {
	// Stopping site also stops the git or package-manager process it is running.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := execute(ctx, osenv.Real{}, os.Args[1:])
	stop()
	os.Exit(code)
}

func execute(ctx context.Context, env cli.Environment, args []string) int {
	s, err := settings.Load(settingsFile, env)
	if err != nil {
		return cli.Fail(env, "site", err)
	}
	return newApp(env, s).Run(ctx, args)
}

// newApp is the one place the concrete types are chosen: the workspace on the settings' folders, and
// git and the package managers run for real.
func newApp(env cli.Environment, s settings.Settings) *cli.App {
	workspace := sites.NewWorkspace(sites.Config{
		Sources:         s.SitesSource,
		Public:          s.SitesPublic,
		Templates:       s.SiteTemplates,
		DefaultTemplate: defaultTemplate,
		Domain:          s.SitesDomain,
	}, run.NewExec(env.Stdout(), env.Stderr()))

	return cli.NewApp("site", usage, env, map[string]cli.Command{
		"templates": sitecmd.NewTemplates(env, workspace),
		"new":       sitecmd.NewNew(env, workspace),
		"preview":   sitecmd.NewPreview(env, workspace),
		"diff":      sitecmd.NewDiff(env, workspace),
		"live":      sitecmd.NewLive(env, workspace),
		"list":      sitecmd.NewList(env, workspace),
	})
}
