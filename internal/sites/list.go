package sites

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Entry is one site as `site list` shows it. A URL is empty when the site has no build at that stage.
type Entry struct {
	Name       string
	PreviewURL string
	LiveURL    string
	Unreleased int // previews committed since the site last went live
}

// List is every site, in name order.
func (w *Workspace) List(ctx context.Context) ([]Entry, error) {
	dirs, err := os.ReadDir(w.cfg.Sources)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var entries []Entry
	for _, d := range dirs {
		// Folders starting with a dot are the agents' own, such as .inbox for files that arrive
		// before their site exists; site names can't start with one.
		if !d.IsDir() || strings.HasPrefix(d.Name(), ".") {
			continue
		}
		entry, err := w.entry(ctx, d.Name())
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (w *Workspace) entry(ctx context.Context, name string) (Entry, error) {
	entry := Entry{Name: name}
	if exists(w.publishedDir("preview", name)) {
		url, err := w.url("preview-" + name)
		if err != nil {
			return Entry{}, err
		}
		entry.PreviewURL = url
	}
	if exists(w.publishedDir("live", name)) {
		url, err := w.url(name)
		if err != nil {
			return Entry{}, err
		}
		entry.LiveURL = url
		entry.Unreleased = w.unreleased(ctx, filepath.Join(w.cfg.Sources, name))
	}
	return entry, nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
