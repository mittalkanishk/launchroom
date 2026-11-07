// Package sites is the launch-room site workspace. Each site's source is a git repository in
// <Sources>/<name> whose first commit is the template, so a diff shows what the room changed. Builds
// are published to <Public>/preview/<name> and <Public>/live/<name>, which the workflows container
// serves at preview-<name>.<domain> and <name>.<domain>.
//
// Read in this order: workspace.go, then one file per step (create, preview, live, diff, list, env),
// then what they share (template, build, publish, history).
package sites

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/mittalkanishk/launch-room/internal/names"
	"github.com/mittalkanishk/launch-room/internal/run"
)

// Runner runs git and the package managers. → internal/run
type Runner interface {
	Run(ctx context.Context, c run.Cmd) error
	Output(ctx context.Context, c run.Cmd) (string, error)
}

// Config is where sites live and how they are built and served.
type Config struct {
	Sources         string // one folder per site's source
	Public          string // published builds: preview/<name> and live/<name>
	Templates       string // one folder per template, each with a launch-room.json
	DefaultTemplate string // the template `site new` uses when none is chosen
	Domain          string // sites are served at <name>.<Domain>
}

type Workspace struct {
	cfg    Config
	runner Runner
}

func NewWorkspace(cfg Config, runner Runner) *Workspace {
	return &Workspace{
		cfg:    cfg,
		runner: runner,
	}
}

// Dir is the source folder of an existing site.
func (w *Workspace) Dir(name string) (string, error) {
	if err := names.Check(name); err != nil {
		return "", err
	}
	dir := filepath.Join(w.cfg.Sources, name)
	if _, err := os.Stat(dir); errors.Is(err, fs.ErrNotExist) {
		return "", fmt.Errorf("no site '%s'; create it with: site new %s", name, name)
	} else if err != nil {
		return "", err
	}
	return dir, nil
}

// url is the public address of host (a site name, or preview-<name>) on the sites domain.
func (w *Workspace) url(host string) (string, error) {
	if w.cfg.Domain == "" {
		return "", errors.New("SITES_DOMAIN is not set")
	}
	return "https://" + host + "." + w.cfg.Domain, nil
}

// publishedDir is the folder a stage ("preview" or "live") of the site is served from.
func (w *Workspace) publishedDir(stage, name string) string {
	return filepath.Join(w.cfg.Public, stage, name)
}
