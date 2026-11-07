package formcmd

import (
	"context"
	"fmt"

	"github.com/mittalkanishk/launch-room/internal/cli"
)

// SignupTester sends a test signup to a site's form. → internal/formspree
type SignupTester interface {
	Test(ctx context.Context, site string) error
}

// TestSignup is `form test <site>`. Each test counts toward Formspree's monthly limit.
type TestSignup struct {
	env   cli.Environment
	forms SignupTester
}

func NewTestSignup(env cli.Environment, forms SignupTester) *TestSignup {
	return &TestSignup{
		env:   env,
		forms: forms,
	}
}

func (c *TestSignup) Run(ctx context.Context, args []string) int {
	if len(args) != 1 {
		return cli.Usage(c.env, "form test <site>")
	}
	if err := c.forms.Test(ctx, args[0]); err != nil {
		return cli.Fail(c.env, prog, err)
	}
	fmt.Fprintln(c.env.Stdout(), "test signup accepted; it shows in the Formspree dashboard")
	return cli.ExitOK
}
