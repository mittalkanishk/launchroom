package roles

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallWritesEachFile(t *testing.T) {
	home := t.TempDir()
	// An earlier start's file is replaced, not appended to.
	stale := filepath.Join(home, ".claude", "CLAUDE.md")
	os.MkdirAll(filepath.Dir(stale), 0o755)
	os.WriteFile(stale, []byte("old role"), 0o644)

	written, err := NewInstaller(home).Install()
	if err != nil {
		t.Fatal(err)
	}
	if len(written) != len(All) {
		t.Fatalf("wrote %d files, want %d", len(written), len(All))
	}

	headings := map[string]string{
		filepath.Join(home, ".claude", "CLAUDE.md"): "# Launch room",
	}
	for path, heading := range headings {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(string(content), heading) {
			t.Errorf("%s starts %q, want %q", path, firstLine(string(content)), heading)
		}
	}

	// Every Claude agent's role is in the shared file.
	claude, _ := os.ReadFile(filepath.Join(home, ".claude", "CLAUDE.md"))
	for _, role := range []string{"If you are `builder`", "If you are `producer`", "If you are `director`", "If you are `editor`", "If you are `reviewer`"} {
		if !strings.Contains(string(claude), role) {
			t.Errorf("CLAUDE.md has no section %q", role)
		}
	}

	// The settings register the gate on every shell command.
	settings, err := os.ReadFile(filepath.Join(home, ".claude", "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	var parsed struct {
		Hooks struct {
			PreToolUse []struct {
				Matcher string `json:"matcher"`
				Hooks   []struct {
					Command string `json:"command"`
				} `json:"hooks"`
			} `json:"PreToolUse"`
		} `json:"hooks"`
	}
	if err := json.Unmarshal(settings, &parsed); err != nil {
		t.Fatalf("settings.json: %v", err)
	}
	hooks := parsed.Hooks.PreToolUse
	if len(hooks) != 1 || hooks[0].Matcher != "Bash" || len(hooks[0].Hooks) != 1 || hooks[0].Hooks[0].Command != "gate" {
		t.Errorf("settings.json hooks = %+v, want one Bash PreToolUse hook running gate", hooks)
	}
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")
	return line
}
