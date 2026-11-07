// Package studio is the service that does the video crew's heavy work: it has explainer scenes
// rendered by their engine's renderer (Manim, or stage for anime.js), joins them, grades and mixes the cut, and measures the result. The
// agents reach it only through the `video` command, over the stack's network.
//
// Read in this order: service.go (accepting a job), steps.go (running each kind), server.go (the
// HTTP API) and client.go (the video command's side).
package studio

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mittalkanishk/launch-room/internal/media"
	"github.com/mittalkanishk/launch-room/internal/names"
	"github.com/mittalkanishk/launch-room/internal/renderer"
)

// Renderer renders a scene's source into a narrated video. → internal/renderer (Client), one per engine
type Renderer interface {
	Render(ctx context.Context, source string) (renderer.Result, error)
}

// Post is the ffmpeg work: joining, colour, loudness and measurement. → internal/media
type Post interface {
	JoinScenes(ctx context.Context, scenes, subtitles []string, burn bool, out string) error
	Finish(ctx context.Context, in, out, look string) error
	Measure(ctx context.Context, in string) (media.Measurements, error)
	Keyframes(ctx context.Context, in, dir string, everySeconds int) ([]string, error)
	Sheet(ctx context.Context, in, subtitles, out string) error
	Motion(ctx context.Context, in, subtitles, out string) error
}

type Service struct {
	dataDir   string
	jobs      *jobStore
	renderers map[string]Renderer // by engine: manim, stage
	post      Post
	slots     chan struct{} // how many jobs run at once
}

func NewService(dataDir string, renderers map[string]Renderer, post Post) (*Service, error) {
	jobs, err := newJobStore(filepath.Join(dataDir, "jobs"))
	if err != nil {
		return nil, err
	}
	return &Service{
		dataDir:   dataDir,
		jobs:      jobs,
		renderers: renderers,
		post:      post,
		slots:     make(chan struct{}, 3),
	}, nil
}

// Submit checks a request and starts it. It returns the queued job at once.
//
// Each job runs in its own goroutine, started here and nowhere else. It waits for one of the
// service's slots, runs its step under a context with the step's time limit, records the result,
// and ends; nothing else is left running.
func (s *Service) Submit(req Request) (Job, error) {
	if err := s.check(req); err != nil {
		return Job{}, err
	}
	job := Job{ID: newID(), Kind: req.Kind, Video: req.Video, Shot: req.Shot, Status: StatusQueued, Created: time.Now().UTC()}
	if err := s.jobs.save(job); err != nil {
		return Job{}, err
	}
	go s.run(job, req)
	return job, nil
}

func (s *Service) Job(id string) (Job, bool) { return s.jobs.get(id) }
func (s *Service) Jobs(video string) []Job   { return s.jobs.list(video) }

// run takes a slot, does the job's step and records how it went.
func (s *Service) run(job Job, req Request) {
	s.slots <- struct{}{}
	defer func() { <-s.slots }()

	job.Status, job.Started = StatusRunning, time.Now().UTC()
	s.jobs.save(job)

	ctx, cancel := context.WithTimeout(context.Background(), timeLimit(req.Kind))
	defer cancel()
	files, err := s.step(ctx, req)
	if err != nil {
		job.Status, job.Message = StatusFailed, err.Error()
	} else {
		job.Status, job.Files, job.Message = StatusDone, files, ""
	}
	s.jobs.save(job)
}

// timeLimit is how long a step may take before it is given up.
func timeLimit(kind string) time.Duration {
	switch kind {
	case KindScene, KindCut:
		return 20 * time.Minute
	default:
		return 5 * time.Minute
	}
}

// videoDir is where a video's renders live in studio's volume.
func (s *Service) videoDir(video string) string { return filepath.Join(s.dataDir, "videos", video) }

// File is the path of one of a video's files, refusing paths that leave the video's folder.
func (s *Service) File(video, rel string) (string, error) {
	if err := names.Check(video); err != nil {
		return "", err
	}
	if !filepath.IsLocal(rel) {
		return "", fmt.Errorf("%q is not a file of %s", rel, video)
	}
	path := filepath.Join(s.videoDir(video), rel)
	if _, err := os.Stat(path); err != nil {
		return "", err
	}
	return path, nil
}

func newID() string {
	b := make([]byte, 6)
	rand.Read(b)
	return time.Now().UTC().Format("20060102-150405-") + hex.EncodeToString(b)
}
