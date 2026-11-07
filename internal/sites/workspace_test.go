package sites

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/mittalkanishk/launch-room/internal/run"
)

// gitAndFakeBuild runs git for real and fakes the package managers and pandoc: a build step writes
// index.html, from the site's src/title.txt, into the build folder its launch-room.json names, and
// pandoc writes the file it is given.
type gitAndFakeBuild struct {
	git   *run.Exec
	steps []string
}

func newRunner(out *bytes.Buffer) *gitAndFakeBuild {
	return &gitAndFakeBuild{git: run.NewExec(out, out)}
}

func (r *gitAndFakeBuild) Run(ctx context.Context, c run.Cmd) error {
	if c.Name == "git" {
		return r.git.Run(ctx, c)
	}
	r.steps = append(r.steps, c.String())
	if c.Name == "pandoc" {
		return os.WriteFile(c.Args[len(c.Args)-1], []byte("pptx"), 0o644)
	}
	if slices.Contains(c.Args, "build") {
		title, err := os.ReadFile(filepath.Join(c.Dir, "src", "title.txt"))
		if err != nil {
			return err
		}
		m, err := readManifest(c.Dir)
		if err != nil {
			return err
		}
		out := filepath.Join(c.Dir, m.buildDir())
		if err := os.MkdirAll(out, 0o755); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(out, "index.html"), title, 0o644)
	}
	return nil
}

func (r *gitAndFakeBuild) Output(ctx context.Context, c run.Cmd) (string, error) {
	return r.git.Output(ctx, c)
}

// newWorkspace is a workspace in a temp folder with two templates. "basic" (the default) builds to
// dist/ and holds a lockfile, plus installed packages and build output a copy must leave behind;
// "report" builds to build/. A folder without launch-room.json is not a template.
func newWorkspace(t *testing.T, out *bytes.Buffer) (*Workspace, *gitAndFakeBuild) {
	t.Helper()
	root := t.TempDir()
	basic := filepath.Join(root, "templates", "basic")
	writeFile(t, filepath.Join(basic, "launch-room.json"), `{"description": "one-page sites", "buildDir": "dist"}`)
	writeFile(t, filepath.Join(basic, "src", "title.txt"), "Your site")
	writeFile(t, filepath.Join(basic, "package-lock.json"), "{}")
	writeFile(t, filepath.Join(basic, "node_modules", "dep", "index.js"), "x")
	writeFile(t, filepath.Join(basic, "dist", "old.html"), "x")

	report := filepath.Join(root, "templates", "report")
	writeFile(t, filepath.Join(report, "launch-room.json"), `{"description": "data reports", "buildDir": "build"}`)
	writeFile(t, filepath.Join(report, "src", "title.txt"), "Report")
	deck := filepath.Join(root, "templates", "deck")
	writeFile(t, filepath.Join(deck, "launch-room.json"), `{"description": "pitch decks", "deck": "slides.md"}`)
	writeFile(t, filepath.Join(deck, "src", "title.txt"), "Deck")
	writeFile(t, filepath.Join(deck, "slides.md"), "# Your company\n")
	writeFile(t, filepath.Join(root, "templates", "notes", "readme.txt"), "not a template")

	runner := newRunner(out)
	w := NewWorkspace(Config{
		Sources:         filepath.Join(root, "sites"),
		Public:          filepath.Join(root, "public"),
		Templates:       filepath.Join(root, "templates"),
		DefaultTemplate: "basic",
		Domain:          "example.test",
	}, runner)
	return w, runner
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func TestWorkflowFromTemplateToLive(t *testing.T) {
	ctx := context.Background()
	var out bytes.Buffer
	w, runner := newWorkspace(t, &out)

	if _, err := w.Create(ctx, "demo", ""); err != nil {
		t.Fatalf("Create: %v\n%s", err, out.String())
	}
	dir := filepath.Join(w.cfg.Sources, "demo")
	if exists(filepath.Join(dir, "node_modules")) || exists(filepath.Join(dir, "dist")) {
		t.Fatal("Create copied the template's node_modules or dist")
	}
	if _, err := w.Create(ctx, "demo", ""); err == nil {
		t.Fatal("Create of an existing site succeeded")
	}

	preview, err := w.Preview(ctx, "demo")
	if err != nil {
		t.Fatalf("Preview: %v\n%s", err, out.String())
	}
	if preview != (Published{URL: "https://preview-demo.example.test"}) {
		t.Errorf("Preview = %+v", preview)
	}
	if want := []string{"npm ci", "npm run build"}; !slices.Equal(runner.steps, want) {
		t.Errorf("build steps = %v, want %v (package-lock.json picks npm ci)", runner.steps, want)
	}

	// An edit that hasn't been previewed can't go live.
	writeFile(t, filepath.Join(dir, "src", "title.txt"), "Quill & Co")
	if _, err := w.Live(ctx, "demo"); err == nil || !strings.Contains(err.Error(), "not in its preview") {
		t.Fatalf("Live with an unpreviewed edit: err = %v", err)
	}

	if _, err := w.Preview(ctx, "demo"); err != nil {
		t.Fatalf("second Preview: %v", err)
	}
	out.Reset()
	if err := w.Diff(ctx, "demo"); err != nil {
		t.Fatalf("Diff: %v", err)
	}
	if !strings.Contains(out.String(), "-Your site") || !strings.Contains(out.String(), "+Quill & Co") {
		t.Errorf("Diff since the template = %q", out.String())
	}

	live, err := w.Live(ctx, "demo")
	if err != nil {
		t.Fatalf("Live: %v", err)
	}
	if live != (Published{URL: "https://demo.example.test"}) {
		t.Errorf("Live = %+v", live)
	}
	if got := readFile(t, filepath.Join(w.publishedDir("live", "demo"), "index.html")); got != "Quill & Co" {
		t.Errorf("live index.html = %q", got)
	}

	entries, err := w.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	want := Entry{Name: "demo", PreviewURL: "https://preview-demo.example.test", LiveURL: "https://demo.example.test"}
	if len(entries) != 1 || entries[0] != want {
		t.Errorf("List = %+v, want [%+v]", entries, want)
	}
}

func TestPreviewNeedsDomainAndExistingSite(t *testing.T) {
	var out bytes.Buffer
	w, _ := newWorkspace(t, &out)

	if _, err := w.Preview(context.Background(), "missing"); err == nil || !strings.Contains(err.Error(), "site new missing") {
		t.Errorf("Preview of a missing site: err = %v", err)
	}
	if _, err := w.Create(context.Background(), "demo", ""); err != nil {
		t.Fatal(err)
	}
	w.cfg.Domain = ""
	if _, err := w.Preview(context.Background(), "demo"); err == nil || !strings.Contains(err.Error(), "SITES_DOMAIN") {
		t.Errorf("Preview without a domain: err = %v", err)
	}
}

func TestPackageManagerFor(t *testing.T) {
	cases := map[string]string{"yarn.lock": "corepack", "pnpm-lock.yaml": "corepack", "package-lock.json": "npm", "": "npm"}
	for lockfile, program := range cases {
		dir := t.TempDir()
		if lockfile != "" {
			writeFile(t, filepath.Join(dir, lockfile), "")
		}
		if got := packageManagerFor(dir); got.program != program || got.lockfile != lockfile {
			t.Errorf("lockfile %q: got %+v", lockfile, got)
		}
	}
}

func TestSetEnvReplacesOnlyItsKey(t *testing.T) {
	var out bytes.Buffer
	w, _ := newWorkspace(t, &out)
	if _, err := w.Create(context.Background(), "demo", ""); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(w.cfg.Sources, "demo", ".env")
	writeFile(t, path, "OTHER=1\nVITE_SIGNUP_ENDPOINT=old\n")

	if err := w.SetEnv("demo", "VITE_SIGNUP_ENDPOINT", "new"); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, path); got != "OTHER=1\nVITE_SIGNUP_ENDPOINT=new\n" {
		t.Errorf(".env = %q", got)
	}
}

func TestTemplatesAndTheirBuildFolders(t *testing.T) {
	ctx := context.Background()
	var out bytes.Buffer
	w, _ := newWorkspace(t, &out)

	templates, err := w.Templates()
	if err != nil {
		t.Fatal(err)
	}
	want := []Template{{ID: "basic", Description: "one-page sites"}, {ID: "deck", Description: "pitch decks"}, {ID: "report", Description: "data reports"}}
	if !slices.Equal(templates, want) {
		t.Errorf("Templates = %+v, want %+v", templates, want)
	}

	if _, err := w.Create(ctx, "q3", "missing"); err == nil || !strings.Contains(err.Error(), "site templates") {
		t.Errorf("Create from a missing template: err = %v", err)
	}
	if _, err := w.Create(ctx, "q3", "../basic"); err == nil {
		t.Error("Create from a template path outside the templates folder succeeded")
	}

	used, err := w.Create(ctx, "q3", "report")
	if err != nil {
		t.Fatalf("Create: %v\n%s", err, out.String())
	}
	if used != "report" {
		t.Errorf("Create used %q, want report", used)
	}
	if _, err := w.Preview(ctx, "q3"); err != nil {
		t.Fatalf("Preview: %v\n%s", err, out.String())
	}
	if got := readFile(t, filepath.Join(w.publishedDir("preview", "q3"), "index.html")); got != "Report" {
		t.Errorf("preview index.html = %q, want the build/ output", got)
	}
}

func TestDeckSitesPublishAnEditablePowerPoint(t *testing.T) {
	ctx := context.Background()
	var out bytes.Buffer
	w, runner := newWorkspace(t, &out)

	if _, err := w.Create(ctx, "pitch", "deck"); err != nil {
		t.Fatal(err)
	}
	preview, err := w.Preview(ctx, "pitch")
	if err != nil {
		t.Fatalf("Preview: %v\n%s", err, out.String())
	}
	if preview.DeckURL != "https://preview-pitch.example.test/deck.pptx" {
		t.Errorf("Preview.DeckURL = %q", preview.DeckURL)
	}
	if got := readFile(t, filepath.Join(w.publishedDir("preview", "pitch"), "deck.pptx")); got != "pptx" {
		t.Errorf("published deck.pptx = %q", got)
	}
	pandoc := "pandoc slides.md --from markdown --to pptx --slide-level 1 --output " + filepath.Join(w.cfg.Sources, "pitch", "dist", "deck.pptx")
	if !slices.Contains(runner.steps, pandoc) {
		t.Errorf("steps = %v\nwant one: %s", runner.steps, pandoc)
	}

	live, err := w.Live(ctx, "pitch")
	if err != nil {
		t.Fatalf("Live: %v", err)
	}
	if live.DeckURL != "https://pitch.example.test/deck.pptx" {
		t.Errorf("Live.DeckURL = %q", live.DeckURL)
	}
}

func TestListSkipsTheAgentsOwnFolders(t *testing.T) {
	var out bytes.Buffer
	w, _ := newWorkspace(t, &out)
	if _, err := w.Create(context.Background(), "demo", ""); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(w.cfg.Sources, ".inbox", "logo.png"), "png")

	entries, err := w.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name != "demo" {
		t.Errorf("List = %+v, want only demo", entries)
	}
}
