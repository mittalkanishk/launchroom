// Package osenv is the real operating-system environment the commands run in. It implements
// cli.Environment and settings.Lookup.
package osenv

import (
	"io"
	"os"
)

type Real struct{}

func (Real) Getenv(key string) string { return os.Getenv(key) }
func (Real) Stdout() io.Writer        { return os.Stdout }
func (Real) Stderr() io.Writer        { return os.Stderr }
