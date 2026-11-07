package videos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/mittalkanishk/launch-room/internal/studio"
)

// renderOrder is the renders from most to least finished.
var renderOrder = []string{"final.mp4", "cut.mp4"}

// publishPreview puts the video's most finished render on a small player page at its preview URL.
func (w *Workspace) publishPreview(name, dir string) (string, error) {
	render := ""
	for _, r := range renderOrder {
		if _, err := os.Stat(filepath.Join(dir, "renders", r)); err == nil {
			render = r
			break
		}
	}
	if render == "" {
		return "", errors.New("there is no render to preview yet")
	}
	p, err := readPlan(dir)
	if err != nil {
		return "", err
	}

	staging := filepath.Join(w.cfg.Public, "preview", "."+name+".new")
	os.RemoveAll(staging)
	if err := os.MkdirAll(staging, 0o755); err != nil {
		return "", err
	}
	if err := copyFile(filepath.Join(dir, "renders", render), filepath.Join(staging, "video.mp4")); err != nil {
		return "", err
	}
	page := playerPage(p.Meta.Title, strings.TrimSuffix(render, ".mp4"))
	if err := os.WriteFile(filepath.Join(staging, "index.html"), []byte(page), 0o644); err != nil {
		return "", err
	}
	// The script page (script.go) lives beside the player; the swap must not lose it.
	target := filepath.Join(w.cfg.Public, "preview", name)
	if _, err := os.Stat(filepath.Join(target, "script")); err == nil {
		if err := copyDir(filepath.Join(target, "script"), filepath.Join(staging, "script")); err != nil {
			return "", err
		}
	}
	if err := swap(staging, target); err != nil {
		return "", err
	}
	return w.url("preview-" + name), nil
}

// Publish puts the final render live, once QC has passed it. Like `site live`, the agents may only
// run it with a person's approval in the room (the gate asks).
func (w *Workspace) Publish(ctx context.Context, name string) (string, error) {
	dir, err := w.Dir(name)
	if err != nil {
		return "", err
	}
	var report studio.Report
	content, err := os.ReadFile(filepath.Join(dir, "renders", "qc", "report.json"))
	if err != nil {
		return "", fmt.Errorf("there is no QC report yet; run: video qc %s", name)
	}
	if err := json.Unmarshal(content, &report); err != nil {
		return "", err
	}
	if report.Render != "final.mp4" {
		return "", fmt.Errorf("QC checked the %s, not the final; run video finish then video qc", report.Render)
	}
	if !report.Passed() {
		return "", fmt.Errorf("QC hasn't passed the final; run: video qc %s", name)
	}
	if _, err := w.publishPreview(name, dir); err != nil {
		return "", err
	}

	staging := filepath.Join(w.cfg.Public, "live", "."+name+".new")
	os.RemoveAll(staging)
	if err := copyDir(filepath.Join(w.cfg.Public, "preview", name), staging); err != nil {
		return "", err
	}
	if err := swap(staging, filepath.Join(w.cfg.Public, "live", name)); err != nil {
		return "", err
	}
	if err := w.commit(ctx, dir, "Published"); err != nil {
		return "", err
	}
	return w.url(name), nil
}

func (w *Workspace) url(host string) string { return "https://" + host + "." + w.cfg.Domain }

// swap replaces target with the staged folder in one rename, so the page is never half-written.
func swap(staging, target string) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	if err := os.RemoveAll(target); err != nil {
		return err
	}
	return os.Rename(staging, target)
}

func copyDir(from, to string) error {
	entries, err := os.ReadDir(from)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(to, 0o755); err != nil {
		return err
	}
	for _, e := range entries {
		if err := copyFile(filepath.Join(from, e.Name()), filepath.Join(to, e.Name())); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(from, to string) error {
	src, err := os.Open(from)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := os.Create(to)
	if err != nil {
		return err
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		return err
	}
	return dst.Close()
}

func playerPage(title, stage string) string {
	t := html.EscapeString(title)
	return `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>` + t + `</title>
<style>
  :root { color-scheme: dark; }
  body { margin: 0; min-height: 100vh; display: grid; place-items: center; gap: 12px;
         background: #0b0b0f; color: #e8e8ee; font-family: system-ui, sans-serif; }
  video { width: min(92vw, 152vh); aspect-ratio: 16 / 9; background: #000; border-radius: 12px; }
  p { margin: 0; font-size: 14px; color: #9a9aa8; }
  a { color: #c7c7ff; }
</style>
</head>
<body>
  <video src="video.mp4" controls playsinline preload="metadata"></video>
  <p>` + t + ` · ` + stage + ` · <a href="video.mp4" download>Download MP4</a></p>
</body>
</html>
`
}
