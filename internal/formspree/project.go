// Package formspree deploys the launch-room sites' signup forms to a Formspree CLI project and sends
// test signups to them. Every site's form is kept in one formspree.json, because each deploy
// disables the project's forms that are missing from the file it deploys.
//
// Read in this order: project.go, then deploy.go (Add), forms.go (Forms), signup.go (Test), and
// config.go, the formspree.json they share.
package formspree

import (
	"context"
	"errors"
	"net/http"

	"github.com/mittalkanishk/launch-room/internal/run"
)

// Runner runs the Formspree CLI. → internal/run
type Runner interface {
	Run(ctx context.Context, c run.Cmd) error
}

// Config is the Formspree CLI project and where its formspree.json is kept.
type Config struct {
	ProjectID string
	DeployKey string
	Dir       string // holds formspree.json
	BaseURL   string // https://formspree.io; tests point it at a local server
}

type Project struct {
	cfg    Config
	runner Runner
	client *http.Client
}

func NewProject(cfg Config, runner Runner, client *http.Client) *Project {
	return &Project{
		cfg:    cfg,
		runner: runner,
		client: client,
	}
}

// Form is one site's signup form.
type Form struct {
	Site     string
	Endpoint string
}

// Endpoint is the address the site's signup form posts to.
func (p *Project) Endpoint(site string) string {
	return p.endpointFor(formKey(site))
}

func (p *Project) endpointFor(key string) string {
	return p.cfg.BaseURL + "/p/" + p.cfg.ProjectID + "/f/" + key
}

// formKey is the site's form in formspree.json, and the last part of its endpoint.
func formKey(site string) string { return site + "-signup" }

func (p *Project) ready() error {
	if p.cfg.ProjectID == "" || p.cfg.DeployKey == "" {
		return errors.New("FORMSPREE_PROJECT_ID and FORMSPREE_DEPLOY_KEY are not set")
	}
	return nil
}
