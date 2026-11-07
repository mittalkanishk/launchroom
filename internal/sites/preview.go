package sites

import (
	"context"
	"path/filepath"
)

// Published is where a site can be seen after a preview or a launch.
type Published struct {
	URL     string
	DeckURL string // the editable PowerPoint, for deck sites only
}

// Preview builds the site, publishes the build at its preview URL and records the site's files as a
// commit.
func (w *Workspace) Preview(ctx context.Context, name string) (Published, error) {
	dir, err := w.Dir(name)
	if err != nil {
		return Published{}, err
	}
	previewURL, err := w.url("preview-" + name)
	if err != nil {
		return Published{}, err
	}

	built, err := w.build(ctx, dir)
	if err != nil {
		return Published{}, err
	}
	if err := w.publish(built, "preview", name); err != nil {
		return Published{}, err
	}
	if err := w.commitChanges(ctx, dir, "Preview"); err != nil {
		return Published{}, err
	}
	return w.published(previewURL, "preview", name), nil
}

// published is the URLs of a site at a stage: its page, and its deck if the stage's build has one.
func (w *Workspace) published(url, stage, name string) Published {
	p := Published{URL: url}
	if exists(filepath.Join(w.publishedDir(stage, name), deckFile)) {
		p.DeckURL = url + "/" + deckFile
	}
	return p
}
