// Package roles gives each agent its launch-room role, by writing the roles' prompts where Claude Code
// reads its standing instructions, ~/.claude/CLAUDE.md, with the settings that run the approval gate
// (cmd/gate) before every shell command. The home is in the data volume, so the files are written on
// every start rather than baked into the image.
package roles

import (
	_ "embed"
	"os"
	"path/filepath"

	"github.com/mittalkanishk/launch-room/internal/prompts"
)

// claudeSettings registers the gate as a Claude Code PreToolUse hook. It replaces ~/.claude/settings.json
// whole, so Claude Code's settings in the controller are only what the launch room sets.
//
//go:embed claude-settings.json
var claudeSettings string

// File is one file an agent reads, with its content and its path relative to the home folder.
type File struct {
	Agent   string
	Path    string
	Content string
}

// All is every file the roles need: every agent is Claude Code, working behind the gate.
var All = []File{
	{Agent: "Claude Code (builder, producer, director, editor, reviewer)", Path: filepath.Join(".claude", "CLAUDE.md"), Content: prompts.Claude},
	{Agent: "Claude Code (approval gate)", Path: filepath.Join(".claude", "settings.json"), Content: claudeSettings},
}

// Installer writes the roles into one home folder.
type Installer struct {
	home string
}

func NewInstaller(home string) *Installer {
	return &Installer{
		home: home,
	}
}

// Install writes every file, replacing what was there. It returns the files it wrote.
func (i *Installer) Install() ([]string, error) {
	var written []string
	for _, f := range All {
		path := filepath.Join(i.home, f.Path)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return written, err
		}
		if err := os.WriteFile(path, []byte(f.Content), 0o644); err != nil {
			return written, err
		}
		written = append(written, path)
	}
	return written, nil
}
