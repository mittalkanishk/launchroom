package sites

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// SetEnv sets key in the site's .env file, which Vite reads when it builds, replacing any earlier
// value and keeping the file's other lines.
func (w *Workspace) SetEnv(name, key, value string) error {
	dir, err := w.Dir(name)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, ".env")

	content, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	var lines []string
	for line := range strings.SplitSeq(strings.TrimRight(string(content), "\n"), "\n") {
		if line != "" && !strings.HasPrefix(line, key+"=") {
			lines = append(lines, line)
		}
	}
	lines = append(lines, key+"="+value)
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644)
}
