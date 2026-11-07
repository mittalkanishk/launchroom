package formspree

import (
	"context"

	"github.com/mittalkanishk/launch-room/internal/names"
	"github.com/mittalkanishk/launch-room/internal/run"
)

// Add puts the site's signup form in formspree.json and deploys the file, which creates the form.
// It returns the form's endpoint.
func (p *Project) Add(ctx context.Context, site string) (string, error) {
	if err := p.ready(); err != nil {
		return "", err
	}
	if err := names.Check(site); err != nil {
		return "", err
	}

	config, err := p.readConfig()
	if err != nil {
		return "", err
	}
	if err := config.setForm(formKey(site), site+" signup"); err != nil {
		return "", err
	}
	if err := p.writeConfig(config); err != nil {
		return "", err
	}

	deploy := run.Cmd{
		Dir:  p.cfg.Dir,
		Name: "formspree",
		Args: []string{"deploy"},
		Env:  []string{"FORMSPREE_DEPLOY_KEY=" + p.cfg.DeployKey},
	}
	if err := p.runner.Run(ctx, deploy); err != nil {
		return "", err
	}
	return p.Endpoint(site), nil
}
