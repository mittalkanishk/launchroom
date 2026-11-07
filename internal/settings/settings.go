// Package settings loads what the site and form commands are configured with. The agents don't
// inherit these settings from the controller, which hands agents only a fixed list of provider
// settings, so the controller's entrypoint writes them to a file. A value set in the environment
// wins over the file.
package settings

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"
)

// Lookup reads one variable from the process environment. → internal/osenv
type Lookup interface {
	Getenv(key string) string
}

type Settings struct {
	SiteTemplates      string // SITE_TEMPLATES: one folder per template new sites are copied from
	SitesDomain        string // SITES_DOMAIN: sites are served at <name>.<domain>
	FormspreeProjectID string // FORMSPREE_PROJECT_ID
	FormspreeDeployKey string // FORMSPREE_DEPLOY_KEY
	OpenAIAPIKey       string // OPENAI_API_KEY: transcribes voice notes
	TranscribeModel    string // TRANSCRIBE_MODEL: the transcription model
	SitesSource        string // SITES_SRC: one folder per site's source
	SitesPublic        string // SITES_PUBLIC: the published builds the workflows container serves
	FormsDir           string // FORMS_DIR: where formspree.json is kept
	StudioURL          string // STUDIO_URL: the video crew's back end
	VideosDir          string // VIDEOS_DIR: one folder per video
}

// fileKeys are the settings the entrypoint writes to the file; anything else in it is ignored.
var fileKeys = map[string]bool{
	"SITE_TEMPLATES":       true,
	"SITES_DOMAIN":         true,
	"FORMSPREE_PROJECT_ID": true,
	"FORMSPREE_DEPLOY_KEY": true,
	"OPENAI_API_KEY":       true,
	"TRANSCRIBE_MODEL":     true,
	"STUDIO_URL":           true,
}

// Load reads the settings file at path (a missing file is fine) and lets env override it.
func Load(path string, env Lookup) (Settings, error) {
	file, err := readFile(path)
	if err != nil {
		return Settings{}, err
	}

	return Settings{
		SiteTemplates:      value(env, file, "SITE_TEMPLATES", "/home/node/templates"),
		SitesDomain:        value(env, file, "SITES_DOMAIN", ""),
		FormspreeProjectID: value(env, file, "FORMSPREE_PROJECT_ID", ""),
		FormspreeDeployKey: value(env, file, "FORMSPREE_DEPLOY_KEY", ""),
		OpenAIAPIKey:       value(env, file, "OPENAI_API_KEY", ""),
		TranscribeModel:    value(env, file, "TRANSCRIBE_MODEL", "gpt-4o-mini-transcribe"),
		SitesSource:        value(env, file, "SITES_SRC", "/home/node/sites"),
		SitesPublic:        value(env, file, "SITES_PUBLIC", "/home/node/public"),
		FormsDir:           value(env, file, "FORMS_DIR", "/home/node/.local/state/launch-room/forms"),
		StudioURL:          value(env, file, "STUDIO_URL", "http://studio:8090"),
		VideosDir:          value(env, file, "VIDEOS_DIR", "/home/node/videos"),
	}, nil
}

// value is key from the environment, else from the file, else fallback.
func value(env Lookup, file map[string]string, key, fallback string) string {
	if v := env.Getenv(key); v != "" {
		return v
	}
	if v := file[key]; v != "" {
		return v
	}
	return fallback
}

// readFile parses KEY=VALUE lines, keeping only fileKeys.
func readFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read settings: %w", err)
	}
	defer f.Close()

	values := map[string]string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		key, val, ok := strings.Cut(scanner.Text(), "=")
		if ok && fileKeys[key] {
			values[key] = val
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read settings: %w", err)
	}
	return values, nil
}
