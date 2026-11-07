package studio

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// The kinds of job studio runs, in the order an explainer goes through them.
const (
	KindScene  = "scene"  // one scene rendered by its engine, its narration and subtitles included
	KindCut    = "cut"    // the scenes joined in order, subtitles burned in unless the request says not to
	KindFinish = "finish" // the cut graded with one look and brought to platform loudness
	KindQC     = "qc"     // measurements and keyframes of the latest render
	KindMotion = "motion" // one rendered scene's motion strip: a row of frames per sentence
)

// Job statuses.
const (
	StatusQueued  = "queued"
	StatusRunning = "running"
	StatusDone    = "done"
	StatusFailed  = "failed"
)

// Request is what the video command asks studio to do.
type Request struct {
	Kind   string `json:"kind"`
	Video  string `json:"video"`
	Shot   int    `json:"shot,omitempty"`   // scene, motion: its number
	Engine string `json:"engine,omitempty"` // scene: manim (a .py scene) or stage (a .js scene)
	Source string `json:"source,omitempty"` // scene: its source
	Shots  []int  `json:"shots,omitempty"`  // cut: the scenes, in order
	// NoSubtitles keeps the subtitles out of the cut's picture; they are still written beside it
	// as cut.srt. Off by default, so an older video command gets them burned in as before.
	NoSubtitles bool   `json:"no_subtitles,omitempty"`
	Look        string `json:"look,omitempty"` // finish: one of media.Looks
}

// Job is one request and how it went. Files are paths inside the video's folder.
type Job struct {
	ID      string    `json:"id"`
	Kind    string    `json:"kind"`
	Video   string    `json:"video"`
	Shot    int       `json:"shot,omitempty"`
	Status  string    `json:"status"`
	Message string    `json:"message,omitempty"` // why it failed
	Files   []string  `json:"files,omitempty"`
	Created time.Time `json:"created"`
	Started time.Time `json:"started,omitempty"` // when it left the queue; zero while queued
	Updated time.Time `json:"updated"`
}

// jobStore keeps every job in memory and as a JSON file, so a restart doesn't lose them.
type jobStore struct {
	dir  string
	mu   sync.Mutex
	jobs map[string]Job
}

func newJobStore(dir string) (*jobStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	s := &jobStore{dir: dir, jobs: map[string]Job{}}
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		var j Job
		if err := json.Unmarshal(content, &j); err != nil {
			return nil, fmt.Errorf("%s: %w", f, err)
		}
		// A job that was running when studio stopped will never finish.
		if j.Status == StatusQueued || j.Status == StatusRunning {
			j.Status, j.Message = StatusFailed, "studio restarted while this job was running; submit it again"
		}
		s.jobs[j.ID] = j
	}
	return s, nil
}

func (s *jobStore) save(j Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	j.Updated = time.Now().UTC()
	s.jobs[j.ID] = j
	content, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.dir, j.ID+".json"), content, 0o644)
}

func (s *jobStore) get(id string) (Job, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, ok := s.jobs[id]
	return j, ok
}

// list is the video's jobs, oldest first.
func (s *jobStore) list(video string) []Job {
	s.mu.Lock()
	defer s.mu.Unlock()
	var jobs []Job
	for _, j := range s.jobs {
		if j.Video == video {
			jobs = append(jobs, j)
		}
	}
	sort.Slice(jobs, func(a, b int) bool { return jobs[a].Created.Before(jobs[b].Created) })
	return jobs
}
