package sites

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// publish swaps the folder from in as the stage's copy of the site, whole: it is copied beside the
// old copy, then renamed over it, so the workflows container never serves half a build.
func (w *Workspace) publish(from, stage, name string) error {
	target := w.publishedDir(stage, name)
	staging := filepath.Join(w.cfg.Public, stage, "."+name+".new")

	if err := os.RemoveAll(staging); err != nil {
		return err
	}
	if err := copyTree(from, staging, nil); err != nil {
		return err
	}
	if err := os.RemoveAll(target); err != nil {
		return err
	}
	return os.Rename(staging, target)
}

// copyTree copies the folder from to the new folder to, keeping file modes and symlinks. Entries at
// the top of from named in skip are left out.
func copyTree(from, to string, skip map[string]bool) error {
	c := treeCopy{from: from, to: to, skip: skip}
	// c.visit is a method value: the ONE place the copy becomes the func WalkDir asks for.
	return filepath.WalkDir(from, c.visit)
}

type treeCopy struct {
	from string
	to   string
	skip map[string]bool
}

func (c treeCopy) visit(path string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(c.from, path)
	if err != nil {
		return err
	}
	if c.skip[rel] {
		if d.IsDir() {
			return filepath.SkipDir
		}
		return nil
	}
	target := filepath.Join(c.to, rel)

	info, err := d.Info()
	if err != nil {
		return err
	}
	switch {
	case d.IsDir():
		return os.MkdirAll(target, info.Mode().Perm()|0o700)
	case d.Type()&fs.ModeSymlink != 0:
		link, err := os.Readlink(path)
		if err != nil {
			return err
		}
		return os.Symlink(link, target)
	default:
		return copyFile(path, target, info.Mode().Perm())
	}
}

func copyFile(from, to string, perm fs.FileMode) error {
	src, err := os.Open(from)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(to, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		return err
	}
	return dst.Close()
}
