package videos

import (
	"context"

	"github.com/mittalkanishk/launch-room/internal/studio"
)

// Checkup is `video check`: whether the plan is ready, and how long it will run.
type Checkup struct {
	Problems      []string
	Scenes        int
	SpokenSeconds int
	TargetSeconds int
}

// Check reads the video's plan and reports its problems and length.
func (w *Workspace) Check(ctx context.Context, name string) (Checkup, error) {
	dir, err := w.Dir(name)
	if err != nil {
		return Checkup{}, err
	}
	p, err := readPlan(dir)
	if err != nil {
		return Checkup{}, err
	}
	return Checkup{
		Problems:      p.Problems(),
		Scenes:        len(p.Scenes),
		SpokenSeconds: p.SpokenSeconds(),
		TargetSeconds: p.Meta.LengthSeconds,
	}, nil
}

// Cut joins the scenes, each as last rendered, with their subtitles burned in unless video.json
// says "subtitles": "none".
func (w *Workspace) Cut(ctx context.Context, name string) (Outcome, error) {
	_, p, err := w.plan(name)
	if err != nil {
		return Outcome{}, err
	}
	var shots []int
	for _, s := range p.Scenes {
		shots = append(shots, s.Shot)
	}
	job, err := w.studio.Submit(ctx, studio.Request{Kind: studio.KindCut, Video: name, Shots: shots, NoSubtitles: !p.Meta.BurnSubtitles()})
	if err != nil {
		return Outcome{}, err
	}
	return w.follow(ctx, name, job)
}

// Finish grades the cut with the look in video.json (or the one given) and brings it to platform
// loudness.
func (w *Workspace) Finish(ctx context.Context, name, look string) (Outcome, error) {
	_, p, err := w.plan(name)
	if err != nil {
		return Outcome{}, err
	}
	if look == "" {
		look = p.Meta.Look
	}
	job, err := w.studio.Submit(ctx, studio.Request{Kind: studio.KindFinish, Video: name, Look: look})
	if err != nil {
		return Outcome{}, err
	}
	return w.follow(ctx, name, job)
}

// QC measures the latest render and fetches its report and keyframes into renders/qc/.
func (w *Workspace) QC(ctx context.Context, name string) (Outcome, error) {
	if _, err := w.Dir(name); err != nil {
		return Outcome{}, err
	}
	job, err := w.studio.Submit(ctx, studio.Request{Kind: studio.KindQC, Video: name})
	if err != nil {
		return Outcome{}, err
	}
	return w.follow(ctx, name, job)
}
