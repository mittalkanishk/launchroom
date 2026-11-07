// Package videos is the video crew's side of the workflow: each video is an explainer, a folder the
// producer, director and editor write (brief, script, references, and one Manim scene per section),
// checked before anything is rendered, and sent to studio one step at a time. Renders come back into
// the folder and are published at preview-<name>.<domain> like the sites.
//
// Read in this order: workspace.go, files.go (the plan and its checks), scenes.go (rendering),
// steps.go (cut, finish, QC), sync.go (waiting and fetching results), publish.go.
package videos

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/mittalkanishk/launch-room/internal/names"
	"github.com/mittalkanishk/launch-room/internal/run"
	"github.com/mittalkanishk/launch-room/internal/studio"
)

//go:embed starter
var starter embed.FS

// Studio does the generation and post-production. → internal/studio (Client)
type Studio interface {
	Submit(ctx context.Context, req studio.Request) (studio.Job, error)
	Job(ctx context.Context, id string) (studio.Job, error)
	Jobs(ctx context.Context, video string) ([]studio.Job, error)
	Download(ctx context.Context, video, rel, out string) error
}

// Runner runs git, which keeps each video's history. → internal/run
type Runner interface {
	Run(ctx context.Context, c run.Cmd) error
	Output(ctx context.Context, c run.Cmd) (string, error)
}

// Config is where videos live and are published.
type Config struct {
	Root        string        // one folder per video
	SiteSources string        // the sites' folder: a video can't take a site's name, they share URLs
	Public      string        // published builds: preview/<name> and live/<name>
	Domain      string        // previews are served at preview-<name>.<Domain>
	Wait        time.Duration // how long a step waits for studio before handing back
	Poll        time.Duration // how often it asks studio while waiting
}

type Workspace struct {
	cfg    Config
	studio Studio
	runner Runner
}

func NewWorkspace(cfg Config, studio Studio, runner Runner) *Workspace {
	return &Workspace{
		cfg:    cfg,
		studio: studio,
		runner: runner,
	}
}

// New creates a video's folder from the starter files, as a git repository.
func (w *Workspace) New(ctx context.Context, name string) (string, error) {
	if err := names.Check(name); err != nil {
		return "", err
	}
	dir := filepath.Join(w.cfg.Root, name)
	for _, taken := range []string{dir, filepath.Join(w.cfg.SiteSources, name)} {
		if _, err := os.Stat(taken); err == nil {
			return "", fmt.Errorf("the name '%s' is taken by a video or a site", name)
		}
	}
	if err := os.MkdirAll(filepath.Join(dir, "references"), 0o755); err != nil {
		return "", err
	}
	if err := copyStarter(dir); err != nil {
		os.RemoveAll(dir)
		return "", err
	}
	if err := w.startHistory(ctx, dir); err != nil {
		os.RemoveAll(dir)
		return "", err
	}
	return dir, nil
}

// Dir is the folder of an existing video.
func (w *Workspace) Dir(name string) (string, error) {
	if err := names.Check(name); err != nil {
		return "", err
	}
	dir := filepath.Join(w.cfg.Root, name)
	if _, err := os.Stat(filepath.Join(dir, "video.json")); errors.Is(err, fs.ErrNotExist) {
		return "", fmt.Errorf("no video '%s'; create it with: video new %s", name, name)
	} else if err != nil {
		return "", err
	}
	return dir, nil
}

// plan reads a video's files and refuses to go on while they have problems.
func (w *Workspace) plan(name string) (string, Plan, error) {
	dir, err := w.Dir(name)
	if err != nil {
		return "", Plan{}, err
	}
	p, err := readPlan(dir)
	if err != nil {
		return "", Plan{}, err
	}
	if problems := p.Problems(); len(problems) > 0 {
		return "", Plan{}, fmt.Errorf("the plan isn't ready (run: video check %s):\n  %s", name, joinLines(problems))
	}
	return dir, p, nil
}

// copyStarter copies the starter files, subfolders included, into dir.
func copyStarter(dir string) error {
	return fs.WalkDir(starter, "starter", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel("starter", path)
		if d.IsDir() {
			return os.MkdirAll(filepath.Join(dir, rel), 0o755)
		}
		content, err := starter.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, rel), content, 0o644)
	})
}

// History: renders stay out of git (they are large and come back from studio); the plan's files
// and references are committed at each step, so the reviewer can see what changed.
func (w *Workspace) startHistory(ctx context.Context, dir string) error {
	if err := w.runner.Run(ctx, gitCmd(dir, "init", "--quiet")); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, ".git", "info", "exclude"), []byte("renders/\n"), 0o644); err != nil {
		return err
	}
	return w.commit(ctx, dir, "Starter")
}

func (w *Workspace) commit(ctx context.Context, dir, message string) error {
	status, err := w.runner.Output(ctx, gitCmd(dir, "status", "--porcelain"))
	if err != nil || status == "" {
		return err
	}
	if err := w.runner.Run(ctx, gitCmd(dir, "add", "--all")); err != nil {
		return err
	}
	return w.runner.Run(ctx, gitCmd(dir, "commit", "--quiet", "-m", message))
}

func gitCmd(dir string, args ...string) run.Cmd {
	identity := []string{"-c", "user.name=video", "-c", "user.email=video@localhost"}
	return run.Cmd{Dir: dir, Name: "git", Args: append(identity, args...)}
}

func joinLines(lines []string) string {
	out := ""
	for i, l := range lines {
		if i > 0 {
			out += "\n  "
		}
		out += l
	}
	return out
}
