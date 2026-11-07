package sites

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/mittalkanishk/launch-room/internal/names"
)

// manifestFile describes a template. It is copied into every site made from the template, so the
// site keeps knowing where its build goes.
const manifestFile = "launch-room.json"

// Template is one choice `site new --template` offers.
type Template struct {
	ID          string
	Description string // what the template is for, in one line
}

// manifest is launch-room.json.
type manifest struct {
	Description string `json:"description"`
	BuildDir    string `json:"buildDir"` // where the build writes its static files; dist if empty
	Deck        string `json:"deck"`     // the Markdown slides of a deck site, exported as deck.pptx
}

// buildDir is where a site built from this manifest writes its static files.
func (m manifest) buildDir() string {
	if m.BuildDir == "" {
		return "dist"
	}
	return m.BuildDir
}

// readManifest reads dir's launch-room.json. A folder without one is a plain site that builds to dist/.
func readManifest(dir string) (manifest, error) {
	var m manifest
	content, err := os.ReadFile(filepath.Join(dir, manifestFile))
	if errors.Is(err, fs.ErrNotExist) {
		return m, nil
	}
	if err != nil {
		return m, err
	}
	if err := json.Unmarshal(content, &m); err != nil {
		return m, fmt.Errorf("%s: %w", filepath.Join(dir, manifestFile), err)
	}
	if !filepath.IsLocal(m.buildDir()) {
		return m, fmt.Errorf("%s: buildDir %q must be a folder inside the site", filepath.Join(dir, manifestFile), m.BuildDir)
	}
	if m.Deck != "" && !filepath.IsLocal(m.Deck) {
		return m, fmt.Errorf("%s: deck %q must be a file inside the site", filepath.Join(dir, manifestFile), m.Deck)
	}
	return m, nil
}

// Templates is every folder in the templates folder that has a launch-room.json, in ID order.
func (w *Workspace) Templates() ([]Template, error) {
	dirs, err := os.ReadDir(w.cfg.Templates)
	if err != nil {
		return nil, err
	}
	var templates []Template
	for _, d := range dirs {
		dir := filepath.Join(w.cfg.Templates, d.Name())
		if !d.IsDir() || !exists(filepath.Join(dir, manifestFile)) {
			continue
		}
		m, err := readManifest(dir)
		if err != nil {
			return nil, err
		}
		templates = append(templates, Template{ID: d.Name(), Description: m.Description})
	}
	return templates, nil
}

// templateSource is where a template choice points: a git URL as given, or else a folder in the
// templates folder. An empty choice is the default template.
func (w *Workspace) templateSource(choice string) (source string, isFolder bool, err error) {
	if choice == "" {
		choice = w.cfg.DefaultTemplate
	}
	if strings.Contains(choice, "://") || strings.HasPrefix(choice, "git@") {
		return choice, false, nil
	}
	if err := names.Check(choice); err != nil {
		return "", false, fmt.Errorf("no template '%s'; run: site templates", choice)
	}
	dir := filepath.Join(w.cfg.Templates, choice)
	if !exists(filepath.Join(dir, manifestFile)) {
		return "", false, fmt.Errorf("no template '%s'; run: site templates", choice)
	}
	return dir, true, nil
}
