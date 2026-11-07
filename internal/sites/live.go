package sites

import (
	"context"
	"fmt"
	"os"
)

// Live puts the site's current preview live and tags that commit as live. It refuses while the site
// has changes its preview doesn't include.
func (w *Workspace) Live(ctx context.Context, name string) (Published, error) {
	dir, err := w.Dir(name)
	if err != nil {
		return Published{}, err
	}
	liveURL, err := w.url(name)
	if err != nil {
		return Published{}, err
	}

	preview := w.publishedDir("preview", name)
	if _, err := os.Stat(preview); err != nil {
		return Published{}, fmt.Errorf("site '%s' has no preview yet; run: site preview %s", name, name)
	}
	clean, err := w.isClean(ctx, dir)
	if err != nil {
		return Published{}, err
	}
	if !clean {
		return Published{}, fmt.Errorf("site '%s' has changes that are not in its preview; run: site preview %s", name, name)
	}

	if err := w.publish(preview, "live", name); err != nil {
		return Published{}, err
	}
	if err := w.tagLive(ctx, dir); err != nil {
		return Published{}, err
	}
	return w.published(liveURL, "live", name), nil
}
