package videocmd

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/mittalkanishk/launch-room/internal/cli"
	"github.com/mittalkanishk/launch-room/internal/studio"
	"github.com/mittalkanishk/launch-room/internal/videos"
)

type fakeTracker struct{ jobs []studio.Job }

func (f fakeTracker) Status(ctx context.Context, name string) ([]studio.Job, []videos.Outcome, error) {
	return f.jobs, nil, nil
}

func TestStatusLeadsWithTheRenderPassAndWhenToPostAgain(t *testing.T) {
	t0 := time.Now().Add(-4 * time.Minute)
	env := &bufferEnv{}
	cmd := NewStatus(env, fakeTracker{jobs: []studio.Job{
		{Kind: studio.KindScene, Shot: 1, Status: studio.StatusDone, Created: t0, Started: t0},
		{Kind: studio.KindScene, Shot: 2, Status: studio.StatusDone, Created: t0, Started: t0},
		{Kind: studio.KindScene, Shot: 3, Status: studio.StatusRunning, Created: t0, Started: t0.Add(2 * time.Minute)},
		{Kind: studio.KindScene, Shot: 4, Status: studio.StatusQueued, Created: t0},
	}})
	if code := cmd.Run(context.Background(), []string{"loop"}); code != cli.ExitOK {
		t.Fatalf("code = %d: %s", code, env.stderr.String())
	}
	lines := strings.Split(strings.TrimSpace(env.stdout.String()), "\n")
	if !strings.HasPrefix(lines[0], "rendering: 2 of 4 scenes done, 1 running, 1 queued; about 4 min to go") {
		t.Errorf("first line = %q, want the pass and its estimate", lines[0])
	}
	if !strings.Contains(lines[1], "every 2 minutes") || !strings.Contains(lines[1], "video status loop") {
		t.Errorf("second line = %q, want when to post the next update", lines[1])
	}
	if !strings.HasPrefix(lines[2], "scene shot 1") {
		t.Errorf("third line = %q, want the job list to follow", lines[2])
	}

	// With nothing rendering there is no pass line, so the room isn't told to keep posting.
	env = &bufferEnv{}
	NewStatus(env, fakeTracker{jobs: []studio.Job{{Kind: studio.KindCut, Status: studio.StatusDone, Created: t0}}}).Run(context.Background(), []string{"loop"})
	if got := env.stdout.String(); strings.Contains(got, "rendering:") || !strings.HasPrefix(got, "cut ") {
		t.Errorf("idle status = %q", got)
	}
}
