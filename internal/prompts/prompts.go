// Package prompts holds the launch-room agents' instructions, compiled into the binaries as strings.
// Edit the Markdown files beside this one; the build embeds them.
//
// The Claude Code agents (builder, producer, director, editor, reviewer) share one home folder, so
// they share one instructions file: Claude puts team.md first, then each role's section, and every
// agent follows the section for its own name.
package prompts

import (
	_ "embed"
	"strings"
)

var (
	//go:embed team.md
	team string
	//go:embed builder.md
	builder string
	//go:embed producer.md
	producer string
	//go:embed director.md
	director string
	//go:embed editor.md
	editor string
	//go:embed reviewer.md
	reviewer string
	//go:embed scenes.md
	scenes string
)

// Claude is the instructions every Claude Code agent reads: the team and how videos are made, then
// one section per role.
var Claude = strings.Join([]string{team, builder, producer, director, editor, reviewer}, "\n---\n\n")

// Scenes is the explainer scene library's reference, printed by `video scenes` for the director.
var Scenes = scenes
