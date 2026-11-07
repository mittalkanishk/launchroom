package sites

import (
	"context"
	"path/filepath"

	"github.com/mittalkanishk/launch-room/internal/run"
)

// deckFile is the editable PowerPoint a deck site publishes beside its web version, so the build
// serves it at <site URL>/deck.pptx.
const deckFile = "deck.pptx"

// exportDeck turns the site's slides (Markdown, named by its launch-room.json) into a PowerPoint
// file in the build folder. Pandoc writes real text boxes, bullets and tables, so the file can be
// edited; the slides are split at level-1 headings, which is how Slidev decks are written.
func (w *Workspace) exportDeck(ctx context.Context, dir string, m manifest, built string) error {
	if m.Deck == "" {
		return nil
	}
	pandoc := run.Cmd{
		Dir:  dir,
		Name: "pandoc",
		Args: []string{m.Deck, "--from", "markdown", "--to", "pptx", "--slide-level", "1", "--output", filepath.Join(built, deckFile)},
	}
	return w.runner.Run(ctx, pandoc)
}
