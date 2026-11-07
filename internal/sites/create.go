package sites

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mittalkanishk/launch-room/internal/names"
	"github.com/mittalkanishk/launch-room/internal/run"
)

// Create copies a template into a new site and records it as the site's first commit. template is
// a template's ID, a git URL, or empty for the default template. Create returns the template it used.
// A site that fails halfway is removed again.
func (w *Workspace) Create(ctx context.Context, name, template string) (string, error) {
	if err := names.Check(name); err != nil {
		return "", err
	}
	source, isFolder, err := w.templateSource(template)
	if err != nil {
		return "", err
	}
	dir := filepath.Join(w.cfg.Sources, name)
	if _, err := os.Stat(dir); err == nil {
		return "", fmt.Errorf("site '%s' already exists", name)
	}

	if err := w.copyTemplate(ctx, source, isFolder, dir); err != nil {
		os.RemoveAll(dir)
		return "", err
	}
	if err := w.startHistory(ctx, dir); err != nil {
		os.RemoveAll(dir)
		return "", err
	}
	return filepath.Base(source), nil
}

// copyTemplate fills dir from the template. A folder (in the repo's templates/, mounted in) is copied
// without its installed packages, build output and git history; a git URL is cloned.
func (w *Workspace) copyTemplate(ctx context.Context, source string, isFolder bool, dir string) error {
	if isFolder {
		m, err := readManifest(source)
		if err != nil {
			return err
		}
		skip := map[string]bool{"node_modules": true, ".git": true, m.buildDir(): true}
		return copyTree(source, dir, skip)
	}

	if err := os.MkdirAll(w.cfg.Sources, 0o755); err != nil {
		return err
	}
	clone := run.Cmd{Name: "git", Args: []string{"clone", "--quiet", "--depth", "1", source, dir}}
	if err := w.runner.Run(ctx, clone); err != nil {
		return err
	}
	return os.RemoveAll(filepath.Join(dir, ".git"))
}
