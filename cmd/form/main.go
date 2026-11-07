// Command form is the launch-room sites' signup forms, collected by Formspree: create a site's form
// and give the site its endpoint, list the forms, and send a test signup.
//
// Read in this order:
//
//	main.go                       which concrete type fills each need (newApp)
//	internal/formcmd              one file per subcommand: arguments in, result out
//	../../internal/formspree      the Formspree project: formspree.json, deploys, test signups
//	../../internal/sites          the site whose .env gets the endpoint
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mittalkanishk/launch-room/cmd/form/internal/formcmd"
	"github.com/mittalkanishk/launch-room/internal/cli"
	"github.com/mittalkanishk/launch-room/internal/formspree"
	"github.com/mittalkanishk/launch-room/internal/osenv"
	"github.com/mittalkanishk/launch-room/internal/run"
	"github.com/mittalkanishk/launch-room/internal/settings"
	"github.com/mittalkanishk/launch-room/internal/sites"
)

// settingsFile is written by the controller's entrypoint; see internal/settings.
const settingsFile = "/home/node/.config/launch-room/settings.env"

const usage = `usage: form add <site>    create the site's signup form and give the site its endpoint
       form list          list the signup forms with their endpoints
       form test <site>   send one test signup to the site's form (counts toward the monthly limit)
`

func main() {
	// Stopping form also stops the Formspree CLI or request it is running.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := execute(ctx, osenv.Real{}, os.Args[1:])
	stop()
	os.Exit(code)
}

func execute(ctx context.Context, env cli.Environment, args []string) int {
	s, err := settings.Load(settingsFile, env)
	if err != nil {
		return cli.Fail(env, "form", err)
	}
	return newApp(env, s).Run(ctx, args)
}

// newApp is the one place the concrete types are chosen: the Formspree project with the real CLI and
// HTTP client, and the site workspace whose .env files get the endpoints.
func newApp(env cli.Environment, s settings.Settings) *cli.App {
	runner := run.NewExec(env.Stdout(), env.Stderr())
	project := formspree.NewProject(formspree.Config{
		ProjectID: s.FormspreeProjectID,
		DeployKey: s.FormspreeDeployKey,
		Dir:       s.FormsDir,
		BaseURL:   "https://formspree.io",
	}, runner, &http.Client{Timeout: 20 * time.Second})
	workspace := sites.NewWorkspace(sites.Config{
		Sources: s.SitesSource,
		Public:  s.SitesPublic,
		Domain:  s.SitesDomain,
	}, runner)

	return cli.NewApp("form", usage, env, map[string]cli.Command{
		"add":  formcmd.NewAdd(env, workspace, project),
		"list": formcmd.NewList(env, project),
		"test": formcmd.NewTestSignup(env, project),
	})
}
