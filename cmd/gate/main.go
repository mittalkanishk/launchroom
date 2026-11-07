// Command gate is the Claude Code PreToolUse hook that stops a site going live without a person's
// approval; see internal/gate. Claude Code pipes the hook input to it and reads its decision back.
// `roles install` registers it in Claude Code's settings.
package main

import (
	"fmt"
	"os"

	"github.com/mittalkanishk/launch-room/internal/gate"
)

func main() {
	if err := gate.Check(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "gate: %v\n", err)
		os.Exit(1)
	}
}
