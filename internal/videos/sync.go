package videos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mittalkanishk/launch-room/internal/studio"
)

// Outcome is where a step got to before the command handed back.
type Outcome struct {
	Job        studio.Job
	Done       bool     // finished, and its files fetched
	Files      []string // the files fetched, in renders/
	PreviewURL string   // set when the step produced a new render to watch
}

// follow waits for one job (up to the workspace's wait), then fetches its files.
func (w *Workspace) follow(ctx context.Context, name string, job studio.Job) (Outcome, error) {
	job, err := w.wait(ctx, job)
	if err != nil {
		return Outcome{Job: job}, err
	}
	if job.Status != studio.StatusDone {
		return Outcome{Job: job}, nil
	}
	return w.fetch(ctx, name, job)
}

// Status lists the video's studio jobs, and fetches the results of any that finished after their
// command handed back.
func (w *Workspace) Status(ctx context.Context, name string) ([]studio.Job, []Outcome, error) {
	dir, err := w.Dir(name)
	if err != nil {
		return nil, nil, err
	}
	jobs, err := w.studio.Jobs(ctx, name)
	if err != nil {
		return nil, nil, err
	}
	fetched := readFetched(dir)
	var outcomes []Outcome
	for _, job := range jobs {
		if job.Status != studio.StatusDone || fetched[job.ID] {
			continue
		}
		outcome, err := w.fetch(ctx, name, job)
		if err != nil {
			return jobs, outcomes, err
		}
		outcomes = append(outcomes, outcome)
	}
	return jobs, outcomes, nil
}

// fetch downloads a finished job's files into renders/, and republishes the preview if it made a
// render.
func (w *Workspace) fetch(ctx context.Context, name string, job studio.Job) (Outcome, error) {
	dir, err := w.Dir(name)
	if err != nil {
		return Outcome{}, err
	}
	out := Outcome{Job: job, Done: true}
	for _, rel := range job.Files {
		local := filepath.Join(dir, "renders", rel)
		if err := w.studio.Download(ctx, name, rel, local); err != nil {
			return out, err
		}
		out.Files = append(out.Files, local)
	}
	if job.Kind == studio.KindCut || job.Kind == studio.KindFinish {
		url, err := w.publishPreview(name, dir)
		if err != nil {
			return out, err
		}
		out.PreviewURL = url
	}
	return out, markFetched(dir, job.ID)
}

// wait polls one job until it is done or failed, or the workspace's wait runs out.
func (w *Workspace) wait(ctx context.Context, job studio.Job) (studio.Job, error) {
	return w.waitUntil(ctx, job, time.Now().Add(w.cfg.Wait))
}

// waitUntil is wait with a deadline shared by several jobs, so a command that follows many
// scenes still hands back on time.
func (w *Workspace) waitUntil(ctx context.Context, job studio.Job, deadline time.Time) (studio.Job, error) {
	for {
		if job.Status == studio.StatusFailed {
			return job, fmt.Errorf("%s failed: %s", job.Kind, job.Message)
		}
		if job.Status == studio.StatusDone || time.Now().After(deadline) {
			return job, nil
		}
		select {
		case <-ctx.Done():
			return job, ctx.Err()
		case <-time.After(w.cfg.Poll):
		}
		next, err := w.studio.Job(ctx, job.ID)
		if err != nil {
			return job, err
		}
		job = next
	}
}

// waitAll waits for every job, and fails if one fails or the wait runs out.
func (w *Workspace) waitAll(ctx context.Context, jobs []studio.Job) ([]studio.Job, error) {
	var done []studio.Job
	var failures []string
	for _, job := range jobs {
		job, err := w.wait(ctx, job)
		switch {
		case err != nil:
			failures = append(failures, fmt.Sprintf("shot %d: %v", job.Shot, err))
		case job.Status != studio.StatusDone:
			failures = append(failures, fmt.Sprintf("shot %d is still %s", job.Shot, job.Status))
		default:
			done = append(done, job)
		}
	}
	if len(failures) > 0 {
		return done, errors.New(strings.Join(failures, "; "))
	}
	return done, nil
}

// fetched records which jobs' files are already in renders/.
const fetchedFile = ".fetched.json"

func readFetched(dir string) map[string]bool {
	ids := map[string]bool{}
	content, err := os.ReadFile(filepath.Join(dir, "renders", fetchedFile))
	if err == nil {
		json.Unmarshal(content, &ids)
	}
	return ids
}

func markFetched(dir, id string) error {
	ids := readFetched(dir)
	ids[id] = true
	content, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dir, "renders"), 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "renders", fetchedFile), content, 0o644)
}
