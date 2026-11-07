package names

import (
	"strings"
	"testing"
)

func TestCheck(t *testing.T) {
	valid := []string{"quillco", "demo-2", "a", strings.Repeat("a", 40)}
	for _, name := range valid {
		if err := Check(name); err != nil {
			t.Errorf("Check(%q) = %v, want nil", name, err)
		}
	}

	invalid := []string{"", "Bad_Name", "-lead", "trail-", "preview-x", "has space", "dot.ted", strings.Repeat("a", 41)}
	for _, name := range invalid {
		if err := Check(name); err == nil {
			t.Errorf("Check(%q) = nil, want an error", name)
		}
	}
}
