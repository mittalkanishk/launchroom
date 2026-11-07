package videos

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mittalkanishk/launch-room/internal/studio"
)

// passWindow is how far apart the scene jobs of one render pass can be submitted. `video render`
// submits a pass within a second; two minutes keeps a retried scene in the same pass.
const passWindow = 2 * time.Minute

// Progress is how a render pass is going: the scene jobs submitted together most recently, with
// an estimate of the time left once a scene has finished. Its String is the line the director
// posts in the room.
type Progress struct {
	Scenes    int           // scenes in the pass
	Done      int           // rendered
	Failed    int           // failed; they count as finished for the estimate
	Running   int           // rendering now
	Queued    int           // waiting for a renderer
	Elapsed   time.Duration // since the first scene of the pass started rendering
	Remaining time.Duration // estimated; zero until a scene has finished
}

// Active reports whether the pass still has scenes rendering or waiting.
func (p Progress) Active() bool { return p.Running+p.Queued > 0 }

// RenderProgress reads the current pass out of a video's jobs: the scene jobs submitted within
// passWindow of the newest one still queued or running. ok is false when nothing is rendering.
func RenderProgress(jobs []studio.Job, now time.Time) (Progress, bool) {
	var newest time.Time
	for _, j := range jobs {
		if j.Kind == studio.KindScene && (j.Status == studio.StatusQueued || j.Status == studio.StatusRunning) && j.Created.After(newest) {
			newest = j.Created
		}
	}
	if newest.IsZero() {
		return Progress{}, false
	}

	var p Progress
	var started time.Time
	for _, j := range jobs {
		if j.Kind != studio.KindScene || j.Created.Before(newest.Add(-passWindow)) || j.Created.After(newest.Add(passWindow)) {
			continue
		}
		p.Scenes++
		switch j.Status {
		case studio.StatusDone:
			p.Done++
		case studio.StatusFailed:
			p.Failed++
		case studio.StatusRunning:
			p.Running++
		default:
			p.Queued++
		}
		if !j.Started.IsZero() && (started.IsZero() || j.Started.Before(started)) {
			started = j.Started
		}
	}
	if started.IsZero() {
		started = newest
	}
	p.Elapsed = now.Sub(started)
	if finished := p.Done + p.Failed; finished > 0 && p.Elapsed > 0 {
		p.Remaining = p.Elapsed / time.Duration(finished) * time.Duration(p.Running+p.Queued)
	}
	return p, true
}

// String is one line for the room: how many scenes are done, running and queued, and the time
// left. The estimate comes from the pass's own pace, so it allows for scenes rendering in parallel.
func (p Progress) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "rendering: %d of %d scenes done", p.Done, p.Scenes)
	if p.Failed > 0 {
		fmt.Fprintf(&b, ", %d failed", p.Failed)
	}
	fmt.Fprintf(&b, ", %d running, %d queued", p.Running, p.Queued)
	switch {
	case !p.Active():
		b.WriteString("; finished")
	case p.Done+p.Failed == 0:
		b.WriteString("; the first scene is still rendering, an estimate follows when it finishes")
	default:
		finished := p.Done + p.Failed
		fmt.Fprintf(&b, "; about %s to go (%s per scene so far)", shortDuration(p.Remaining), shortDuration(p.Elapsed/time.Duration(finished)))
	}
	return b.String()
}

// shortDuration is "45 s", "1 min 40 s" or "9 min", rounded to the second.
func shortDuration(d time.Duration) string {
	d = d.Round(time.Second)
	if d < time.Minute {
		return fmt.Sprintf("%d s", int(d.Seconds()))
	}
	m, s := int(d.Minutes()), int(d.Seconds())%60
	if s == 0 || m >= 5 {
		return fmt.Sprintf("%d min", m+boolInt(s >= 30 && m >= 5))
	}
	return fmt.Sprintf("%d min %d s", m, s)
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Progress reports the video's current render pass, from studio's jobs. ok is false when nothing
// is rendering.
func (w *Workspace) Progress(ctx context.Context, name string) (Progress, bool, error) {
	if _, err := w.Dir(name); err != nil {
		return Progress{}, false, err
	}
	jobs, err := w.studio.Jobs(ctx, name)
	if err != nil {
		return Progress{}, false, err
	}
	p, ok := RenderProgress(jobs, time.Now())
	return p, ok, nil
}
