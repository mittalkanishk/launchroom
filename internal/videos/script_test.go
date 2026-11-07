package videos

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTheScriptPreviewIsAPageOfSectionsWithTheirLengths(t *testing.T) {
	w := newTestWorkspace(t, fakePost{})
	ctx := context.Background()
	dir, err := w.New(ctx, "loop")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.ScriptPreview(ctx, "loop"); err == nil {
		t.Fatal("the starter's empty script was previewed")
	}

	long := strings.Repeat("Agents run in a loop. ", 12) // 60 words, in range
	script := "## 1. The loop\n" + long + "\n\n## 2. The *code*\nHere it is, with `go run`.\n\n| a | b |\n| - | - |\n| 1 | 2 |\n"
	if err := os.WriteFile(filepath.Join(dir, "script.md"), []byte(script), 0o644); err != nil {
		t.Fatal(err)
	}
	url, err := w.ScriptPreview(ctx, "loop")
	if err != nil || url != "https://preview-loop.example.test/script/" {
		t.Fatalf("preview = %q, %v", url, err)
	}
	page, err := os.ReadFile(filepath.Join(w.cfg.Public, "preview", "loop", "script", "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"<h1>loop</h1>", // the starter has no title yet, so the name
		"<h2>1. The loop</h2>",
		`<p class="note">60 words · about 20s</p>`,
		"<h2>2. The <em>code</em></h2>",
		`<p class="note warn">10 words · about 3s</p>`, // under 40 words; the table's pipes don't count
		"<code>go run</code>",
		"<table>",
		`<p class="note warn">70 words, about 23s spoken</p>`, // far from the starter's 240s
	} {
		if !strings.Contains(string(page), want) {
			t.Errorf("page lacks %q:\n%s", want, page)
		}
	}
}

func TestTheScriptPageSurvivesTheFirstCut(t *testing.T) {
	w := newTestWorkspace(t, fakePost{})
	ctx := context.Background()
	dir, err := w.New(ctx, "loop")
	if err != nil {
		t.Fatal(err)
	}
	fillPlan(t, dir)
	if _, err := w.ScriptPreview(ctx, "loop"); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Render(ctx, "loop", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Cut(ctx, "loop"); err != nil {
		t.Fatal(err)
	}
	preview := filepath.Join(w.cfg.Public, "preview", "loop")
	for _, f := range []string{"index.html", "video.mp4", "script/index.html"} {
		if _, err := os.Stat(filepath.Join(preview, f)); err != nil {
			t.Errorf("after the cut, the preview lacks %s", f)
		}
	}
}
