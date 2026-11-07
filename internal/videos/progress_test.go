package videos

import (
	"testing"
	"time"

	"github.com/mittalkanishk/launch-room/internal/studio"
)

func TestRenderProgressFollowsTheLatestPassAndEstimatesFromItsPace(t *testing.T) {
	t0 := time.Date(2026, 11, 7, 0, 0, 0, 0, time.UTC)
	scene := func(shot int, status string, created, started time.Time) studio.Job {
		return studio.Job{Kind: studio.KindScene, Shot: shot, Status: status, Created: created, Started: started}
	}
	old := t0.Add(-time.Hour) // an earlier pass, all done: not part of the current one
	jobs := []studio.Job{
		scene(1, studio.StatusDone, old, old),
		scene(2, studio.StatusDone, old, old),
		{Kind: studio.KindCut, Status: studio.StatusDone, Created: old.Add(10 * time.Minute)},
		scene(1, studio.StatusDone, t0, t0.Add(time.Second)),
		scene(2, studio.StatusDone, t0, t0.Add(time.Second)),
		scene(3, studio.StatusFailed, t0, t0.Add(90*time.Second)),
		scene(4, studio.StatusRunning, t0, t0.Add(3*time.Minute)),
		scene(5, studio.StatusQueued, t0, time.Time{}),
		scene(6, studio.StatusQueued, t0, time.Time{}),
	}

	now := t0.Add(time.Second + 6*time.Minute) // three finished in six minutes: two minutes each
	p, ok := RenderProgress(jobs, now)
	if !ok {
		t.Fatal("a pass with queued scenes isn't active")
	}
	want := Progress{Scenes: 6, Done: 2, Failed: 1, Running: 1, Queued: 2, Elapsed: 6 * time.Minute, Remaining: 6 * time.Minute}
	if p != want {
		t.Errorf("progress = %+v, want %+v", p, want)
	}
	if got := p.String(); got != "rendering: 2 of 6 scenes done, 1 failed, 1 running, 2 queued; about 6 min to go (2 min per scene so far)" {
		t.Errorf("line = %q", got)
	}

	if _, ok := RenderProgress(jobs[:3], now); ok {
		t.Error("a video with nothing rendering reported a pass")
	}
}

func TestRenderProgressHasNoEstimateBeforeTheFirstSceneFinishes(t *testing.T) {
	t0 := time.Now().Add(-40 * time.Second)
	jobs := []studio.Job{
		{Kind: studio.KindScene, Shot: 1, Status: studio.StatusRunning, Created: t0, Started: t0},
		{Kind: studio.KindScene, Shot: 2, Status: studio.StatusQueued, Created: t0},
	}
	p, ok := RenderProgress(jobs, time.Now())
	if !ok || p.Remaining != 0 || p.Scenes != 2 {
		t.Fatalf("progress = %+v, %v", p, ok)
	}
	if got := p.String(); got != "rendering: 0 of 2 scenes done, 1 running, 1 queued; the first scene is still rendering, an estimate follows when it finishes" {
		t.Errorf("line = %q", got)
	}
}

func TestShortDuration(t *testing.T) {
	for d, want := range map[time.Duration]string{
		45 * time.Second:                "45 s",
		100 * time.Second:               "1 min 40 s",
		3 * time.Minute:                 "3 min",
		9*time.Minute + 20*time.Second:  "9 min",
		19*time.Minute + 40*time.Second: "20 min",
	} {
		if got := shortDuration(d); got != want {
			t.Errorf("shortDuration(%s) = %q, want %q", d, got, want)
		}
	}
}
