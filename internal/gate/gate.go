// Package gate is the approval gate on the steps a person must approve: putting a site live and
// publishing a video. Claude Code runs it as a PreToolUse hook
// before every shell command; when the command is one of those steps, the gate answers "ask", and
// Claude Code asks permission instead of running it. In the controller that permission request is
// an approval prompt in the room, with buttons only a person can press. So none of these happen on
// an agent's say-so, whatever its instructions say.
package gate

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
)

// Gated is one step that needs a person's approval, and what the room is told on the prompt (and
// the agent, if it is declined).
type Gated struct {
	Command string // the words that run it: "site live"
	Reason  string
}

// Steps are every gated step.
var Steps = []Gated{
	{Command: "site live", Reason: "Going live needs a person in the room to approve it."},
	{Command: "video publish", Reason: "Publishing a video needs a person in the room to approve it."},
}

// Input is the part of Claude Code's PreToolUse hook input the gate reads.
type Input struct {
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		Command string `json:"command"`
	} `json:"tool_input"`
}

// askOutput is the hook output that makes Claude Code ask permission for the tool call.
type askOutput struct {
	HookSpecificOutput struct {
		HookEventName            string `json:"hookEventName"`
		PermissionDecision       string `json:"permissionDecision"`
		PermissionDecisionReason string `json:"permissionDecisionReason"`
	} `json:"hookSpecificOutput"`
}

// patterns match each step anywhere a shell would run it: at the start, after a separator
// (; & | newline), in a subshell or quote, or by its full path.
var patterns = compile(Steps)

func compile(steps []Gated) []*regexp.Regexp {
	boundary := `(^|[\s;&|(/"'` + "`" + `])`
	out := make([]*regexp.Regexp, len(steps))
	for i, s := range steps {
		words := regexp.MustCompile(`\s+`).ReplaceAllString(regexp.QuoteMeta(s.Command), `\s+`)
		out[i] = regexp.MustCompile(boundary + words + `(\s|$)`)
	}
	return out
}

// Match is the gated step a shell command runs, if any.
func Match(command string) (Gated, bool) {
	for i, p := range patterns {
		if p.MatchString(command) {
			return Steps[i], true
		}
	}
	return Gated{}, false
}

// GoesLive reports whether a shell command would put a site live.
func GoesLive(command string) bool {
	step, ok := Match(command)
	return ok && step.Command == "site live"
}

// Check reads one hook input and writes the "ask" decision if the tool call is a gated step. Any
// other tool call gets no output, which lets Claude Code decide as it normally would.
func Check(in io.Reader, out io.Writer) error {
	var input Input
	if err := json.NewDecoder(in).Decode(&input); err != nil {
		return fmt.Errorf("hook input: %w", err)
	}
	if input.ToolName != "Bash" {
		return nil
	}
	step, ok := Match(input.ToolInput.Command)
	if !ok {
		return nil
	}

	var ask askOutput
	ask.HookSpecificOutput.HookEventName = "PreToolUse"
	ask.HookSpecificOutput.PermissionDecision = "ask"
	ask.HookSpecificOutput.PermissionDecisionReason = step.Reason
	return json.NewEncoder(out).Encode(ask)
}
