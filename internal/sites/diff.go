package sites

import (
	"context"

	"github.com/mittalkanishk/launch-room/internal/run"
)

// Diff writes what changed in the site since it last went live, or since the template if it never
// has, including changes not yet previewed.
func (w *Workspace) Diff(ctx context.Context, name string) error {
	dir, err := w.Dir(name)
	if err != nil {
		return err
	}

	base := liveTag
	if !w.hasLiveTag(ctx, dir) {
		if base, err = w.firstCommit(ctx, dir); err != nil {
			return err
		}
	}
	return w.runner.Run(ctx, run.Cmd{Dir: dir, Name: "git", Args: []string{"--no-pager", "diff", base}})
}
