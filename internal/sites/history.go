package sites

import (
	"context"
	"os"
	"path/filepath"
	"strconv"

	"github.com/mittalkanishk/launch-room/internal/run"
)

// A site's git history: the template is the first commit, each preview with changes is a commit,
// and the tag liveTag marks the commit that is live.
const liveTag = "live"

// gitCmd runs git in dir as the "site" author, since the container has no git identity.
func gitCmd(dir string, args ...string) run.Cmd {
	identity := []string{"-c", "user.name=site", "-c", "user.email=site@localhost"}
	return run.Cmd{Dir: dir, Name: "git", Args: append(identity, args...)}
}

// startHistory makes dir a repository whose first commit is the template, leaving the installed
// packages and the build output out of it.
func (w *Workspace) startHistory(ctx context.Context, dir string) error {
	m, err := readManifest(dir)
	if err != nil {
		return err
	}
	if err := w.runner.Run(ctx, gitCmd(dir, "init", "--quiet")); err != nil {
		return err
	}
	exclude := "node_modules/\n" + m.buildDir() + "/\n"
	if err := os.WriteFile(filepath.Join(dir, ".git", "info", "exclude"), []byte(exclude), 0o644); err != nil {
		return err
	}
	if err := w.runner.Run(ctx, gitCmd(dir, "add", "--all")); err != nil {
		return err
	}
	return w.runner.Run(ctx, gitCmd(dir, "commit", "--quiet", "-m", "Template"))
}

// commitChanges records the site's files as a commit, if they changed since the last one.
func (w *Workspace) commitChanges(ctx context.Context, dir, message string) error {
	clean, err := w.isClean(ctx, dir)
	if err != nil || clean {
		return err
	}
	if err := w.runner.Run(ctx, gitCmd(dir, "add", "--all")); err != nil {
		return err
	}
	return w.runner.Run(ctx, gitCmd(dir, "commit", "--quiet", "-m", message))
}

// isClean reports whether the site's files match its last commit.
func (w *Workspace) isClean(ctx context.Context, dir string) (bool, error) {
	status, err := w.runner.Output(ctx, gitCmd(dir, "status", "--porcelain"))
	return status == "", err
}

func (w *Workspace) tagLive(ctx context.Context, dir string) error {
	return w.runner.Run(ctx, gitCmd(dir, "tag", "--force", liveTag))
}

func (w *Workspace) hasLiveTag(ctx context.Context, dir string) bool {
	_, err := w.runner.Output(ctx, gitCmd(dir, "rev-parse", "--quiet", "--verify", liveTag))
	return err == nil
}

func (w *Workspace) firstCommit(ctx context.Context, dir string) (string, error) {
	return w.runner.Output(ctx, gitCmd(dir, "rev-list", "--max-parents=0", "HEAD"))
}

// unreleased is how many commits the site has since it last went live; 0 if that can't be told.
func (w *Workspace) unreleased(ctx context.Context, dir string) int {
	out, err := w.runner.Output(ctx, gitCmd(dir, "rev-list", "--count", liveTag+"..HEAD"))
	if err != nil {
		return 0
	}
	n, _ := strconv.Atoi(out)
	return n
}
