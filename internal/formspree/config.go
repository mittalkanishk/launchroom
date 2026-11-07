package formspree

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
)

// configFile is formspree.json: {"forms": {"<key>": {"name": ...}}}. Anything this package doesn't
// set, such as actions added to a form by hand, is kept as it is.
type configFile struct {
	root  map[string]json.RawMessage
	forms map[string]json.RawMessage
}

func (p *Project) configPath() string { return filepath.Join(p.cfg.Dir, "formspree.json") }

// readConfig reads formspree.json; a missing file is an empty one.
func (p *Project) readConfig() (*configFile, error) {
	c := &configFile{root: map[string]json.RawMessage{}, forms: map[string]json.RawMessage{}}

	content, err := os.ReadFile(p.configPath())
	if errors.Is(err, fs.ErrNotExist) {
		return c, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(content, &c.root); err != nil {
		return nil, fmt.Errorf("%s: %w", p.configPath(), err)
	}
	if forms, ok := c.root["forms"]; ok {
		if err := json.Unmarshal(forms, &c.forms); err != nil {
			return nil, fmt.Errorf("%s: forms: %w", p.configPath(), err)
		}
	}
	return c, nil
}

func (p *Project) writeConfig(c *configFile) error {
	forms, err := json.Marshal(c.forms)
	if err != nil {
		return err
	}
	c.root["forms"] = forms

	content, err := json.MarshalIndent(c.root, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(p.cfg.Dir, 0o700); err != nil {
		return err
	}
	return os.WriteFile(p.configPath(), append(content, '\n'), 0o600)
}

// setForm adds the form under key, or replaces it.
func (c *configFile) setForm(key, name string) error {
	form, err := json.Marshal(map[string]string{"name": name})
	if err != nil {
		return err
	}
	c.forms[key] = form
	return nil
}

func (c *configFile) formKeys() []string {
	keys := make([]string, 0, len(c.forms))
	for key := range c.forms {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
