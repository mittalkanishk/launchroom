package formspree

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/mittalkanishk/launch-room/internal/run"
)

type recordingRunner struct {
	cmds []run.Cmd
}

func (r *recordingRunner) Run(ctx context.Context, c run.Cmd) error {
	r.cmds = append(r.cmds, c)
	return nil
}

func newProject(t *testing.T, baseURL string) (*Project, *recordingRunner) {
	t.Helper()
	runner := &recordingRunner{}
	p := NewProject(Config{
		ProjectID: "proj123",
		DeployKey: "deploy-key",
		Dir:       t.TempDir(),
		BaseURL:   baseURL,
	}, runner, http.DefaultClient)
	return p, runner
}

func TestAddKeepsOtherFormsAndDeploys(t *testing.T) {
	p, runner := newProject(t, "https://formspree.example")
	existing := `{"forms":{"old-signup":{"name":"old signup","actions":[{"app":"email"}]}}}`
	if err := os.WriteFile(p.configPath(), []byte(existing), 0o600); err != nil {
		t.Fatal(err)
	}

	endpoint, err := p.Add(context.Background(), "quillco")
	if err != nil {
		t.Fatal(err)
	}
	if endpoint != "https://formspree.example/p/proj123/f/quillco-signup" {
		t.Errorf("endpoint = %q", endpoint)
	}

	var saved struct {
		Forms map[string]map[string]any `json:"forms"`
	}
	content, _ := os.ReadFile(filepath.Join(p.cfg.Dir, "formspree.json"))
	if err := json.Unmarshal(content, &saved); err != nil {
		t.Fatal(err)
	}
	if saved.Forms["quillco-signup"]["name"] != "quillco signup" {
		t.Errorf("new form = %v", saved.Forms["quillco-signup"])
	}
	if _, ok := saved.Forms["old-signup"]["actions"]; !ok {
		t.Errorf("the existing form lost its actions: %v", saved.Forms["old-signup"])
	}

	if len(runner.cmds) != 1 {
		t.Fatalf("ran %d commands, want 1 deploy", len(runner.cmds))
	}
	deploy := runner.cmds[0]
	if deploy.String() != "formspree deploy" || deploy.Dir != p.cfg.Dir || !slices.Contains(deploy.Env, "FORMSPREE_DEPLOY_KEY=deploy-key") {
		t.Errorf("deploy = %+v", deploy)
	}
}

func TestFormsListsEverySite(t *testing.T) {
	p, _ := newProject(t, "https://formspree.example")
	for _, site := range []string{"zeta", "alpha"} {
		if _, err := p.Add(context.Background(), site); err != nil {
			t.Fatal(err)
		}
	}

	forms, err := p.Forms()
	if err != nil {
		t.Fatal(err)
	}
	want := []Form{
		{Site: "alpha", Endpoint: "https://formspree.example/p/proj123/f/alpha-signup"},
		{Site: "zeta", Endpoint: "https://formspree.example/p/proj123/f/zeta-signup"},
	}
	if !slices.Equal(forms, want) {
		t.Errorf("Forms = %+v, want %+v", forms, want)
	}
}

func TestTestSignup(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if r.URL.Path == "/p/proj123/f/good-signup" && body["email"] == "form-test+good@example.com" && r.Header.Get("Accept") == "application/json" {
			w.Write([]byte(`{"ok":true,"next":"/thanks"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Form not found"}`))
	}))
	defer server.Close()
	p, _ := newProject(t, server.URL)

	if err := p.Test(context.Background(), "good"); err != nil {
		t.Errorf("Test(good) = %v", err)
	}
	if err := p.Test(context.Background(), "missing"); err == nil || !strings.Contains(err.Error(), "Form not found") {
		t.Errorf("Test(missing) = %v, want Formspree's error", err)
	}
}

func TestNotConfigured(t *testing.T) {
	p := NewProject(Config{Dir: t.TempDir()}, &recordingRunner{}, http.DefaultClient)
	if _, err := p.Add(context.Background(), "demo"); err == nil || !strings.Contains(err.Error(), "FORMSPREE_PROJECT_ID") {
		t.Errorf("Add without a project: err = %v", err)
	}
}
