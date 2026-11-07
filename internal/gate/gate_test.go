package gate

import (
	"bytes"
	"strings"
	"testing"
)

func TestGoesLive(t *testing.T) {
	live := []string{
		"site live quillco",
		"site  live quillco",
		"/usr/local/bin/site live quillco",
		"cd ~/sites/quillco && site live quillco",
		"site preview quillco; site live quillco",
		"site preview quillco | site live quillco",
		`sh -c "site live quillco"`,
		"(site live quillco)",
		"site live",
	}
	for _, c := range live {
		if !GoesLive(c) {
			t.Errorf("GoesLive(%q) = false, want true", c)
		}
	}

	notLive := []string{
		"site list",
		"site preview quillco",
		"site diff quillco",
		"website live quillco",
		"echo 'the site livestream'",
		"form add quillco",
		"",
	}
	for _, c := range notLive {
		if GoesLive(c) {
			t.Errorf("GoesLive(%q) = true, want false", c)
		}
	}
}

func TestCheckAsksOnlyForGoingLive(t *testing.T) {
	cases := []struct {
		input string
		ask   bool
	}{
		{`{"tool_name":"Bash","tool_input":{"command":"site live quillco"}}`, true},
		{`{"tool_name":"Bash","tool_input":{"command":"site preview quillco"}}`, false},
		{`{"tool_name":"Write","tool_input":{"file_path":"x","content":"site live quillco"}}`, false},
	}
	for _, c := range cases {
		var out bytes.Buffer
		if err := Check(strings.NewReader(c.input), &out); err != nil {
			t.Fatalf("Check(%s): %v", c.input, err)
		}
		if got := strings.Contains(out.String(), `"permissionDecision":"ask"`); got != c.ask {
			t.Errorf("Check(%s) output %q; asked = %v, want %v", c.input, out.String(), got, c.ask)
		}
	}

	if err := Check(strings.NewReader("not json"), &bytes.Buffer{}); err == nil {
		t.Error("Check with bad input: err = nil, want an error")
	}
}

func TestVideoStepsAreGated(t *testing.T) {
	gated := map[string]string{
		"video publish promo":                      "video publish",
		"/usr/local/bin/video publish promo":       "video publish",
		"cd ~/videos/promo && video publish promo": "video publish",
	}
	for command, want := range gated {
		step, ok := Match(command)
		if !ok || step.Command != want {
			t.Errorf("Match(%q) = %q, %v; want %q", command, step.Command, ok, want)
		}
	}
	for _, free := range []string{"video render promo", "video cut promo", "video status promo", "video check promo"} {
		if step, ok := Match(free); ok {
			t.Errorf("Match(%q) gated it as %q", free, step.Command)
		}
	}
}
