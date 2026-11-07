// Package names holds the rules for site names, which become hostnames and folder names.
package names

import (
	"fmt"
	"strings"
)

const maxLength = 40

// Check reports whether name can be a site: lowercase letters, digits and hyphens, not starting or
// ending with a hyphen, and not starting with "preview-", which marks preview hostnames.
func Check(name string) error {
	if len(name) > maxLength {
		return fmt.Errorf("name '%s' is longer than %d characters", name, maxLength)
	}
	if name == "" || strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") ||
		strings.HasPrefix(name, "preview-") || !onlyAllowed(name) {
		return fmt.Errorf("invalid name '%s': use lowercase letters, digits and hyphens, not starting with preview-", name)
	}
	return nil
}

func onlyAllowed(name string) bool {
	for _, r := range name {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-') {
			return false
		}
	}
	return true
}
