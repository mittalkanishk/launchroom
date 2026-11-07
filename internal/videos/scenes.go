package videos

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/mittalkanishk/launch-room/internal/studio"
)

// Scene is one file in scenes/: its number, its name, the engine that renders it, and its source.
// The director writes one per section of the script, with the narration inside: a .js file is
// animated with anime.js (the stage engine), a .py file with Manim.
type Scene struct {
	Shot   int
	File   string
	Engine string
	Source string
}

// sceneFile is "01.js", "01-the-loop.js" or the same with .py.
var sceneFile = regexp.MustCompile(`^(\d{2})(-[a-z0-9-]+)?\.(js|py)$`)

var engines = map[string]string{"js": "stage", "py": "manim"}

// readScenes reads scenes/ in order.
func readScenes(dir string) ([]Scene, error) {
	entries, err := os.ReadDir(filepath.Join(dir, "scenes"))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var scenes []Scene
	for _, e := range entries {
		m := sceneFile.FindStringSubmatch(e.Name())
		if e.IsDir() || m == nil {
			continue
		}
		source, err := os.ReadFile(filepath.Join(dir, "scenes", e.Name()))
		if err != nil {
			return nil, err
		}
		shot, _ := strconv.Atoi(m[1])
		scenes = append(scenes, Scene{Shot: shot, File: e.Name(), Engine: engines[m[3]], Source: string(source)})
	}
	sort.Slice(scenes, func(a, b int) bool { return scenes[a].Shot < scenes[b].Shot })
	return scenes, nil
}

// SceneResult is how one scene's render (or motion strip) went.
type SceneResult struct {
	Shot    int
	File    string // the scene's source file for a render; the strip's path for a motion strip
	Status  string // done, failed, or still rendering
	Message string // for a failure, Manim's error naming the line
}

// Render renders the scenes asked for (all if none) at full HD. It waits for them up to the
// workspace's wait and brings back what finished (the video, subtitles and sheet of each scene,
// into renders/scenes/); scenes still rendering carry on, and `video status` fetches them.
func (w *Workspace) Render(ctx context.Context, name string, shots []int) ([]SceneResult, error) {
	dir, p, err := w.plan(name)
	if err != nil {
		return nil, err
	}
	var jobs []studio.Job
	files := map[int]string{}
	for _, s := range p.pick(shots) {
		job, err := w.studio.Submit(ctx, studio.Request{Kind: studio.KindScene, Video: name, Shot: s.Shot, Engine: s.Engine, Source: s.Source})
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
		files[s.Shot] = s.File
	}
	var results []SceneResult
	deadline := time.Now().Add(w.cfg.Wait)
	for _, job := range jobs {
		job, _ = w.waitUntil(ctx, job, deadline)
		status := job.Status
		if status == studio.StatusQueued || status == studio.StatusRunning {
			status = "still rendering"
		}
		if status == studio.StatusDone {
			if _, err := w.fetch(ctx, name, job); err != nil {
				return results, err
			}
		}
		results = append(results, SceneResult{Shot: job.Shot, File: files[job.Shot], Status: status, Message: job.Message})
	}
	return results, w.commit(ctx, dir, "Render")
}

// Motion asks studio for the motion strip of each scene asked for (all if none): one row of
// frames per sentence, into renders/scenes/shot-NN-motion.jpg, for judging how things move. The
// scenes must be rendered already. Like Render, it waits up to the workspace's wait.
func (w *Workspace) Motion(ctx context.Context, name string, shots []int) ([]SceneResult, error) {
	_, p, err := w.plan(name)
	if err != nil {
		return nil, err
	}
	var jobs []studio.Job
	for _, s := range p.pick(shots) {
		job, err := w.studio.Submit(ctx, studio.Request{Kind: studio.KindMotion, Video: name, Shot: s.Shot})
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	var results []SceneResult
	deadline := time.Now().Add(w.cfg.Wait)
	for _, job := range jobs {
		job, _ = w.waitUntil(ctx, job, deadline)
		status := job.Status
		if status == studio.StatusQueued || status == studio.StatusRunning {
			status = "still rendering"
		}
		result := SceneResult{Shot: job.Shot, Status: status, Message: job.Message}
		if status == studio.StatusDone {
			out, err := w.fetch(ctx, name, job)
			if err != nil {
				return results, err
			}
			if len(out.Files) > 0 {
				result.File = out.Files[0]
			}
		}
		results = append(results, result)
	}
	return results, nil
}

// pick is the scenes asked for by number, or all of them.
func (p Plan) pick(numbers []int) []Scene {
	if len(numbers) == 0 {
		return p.Scenes
	}
	var out []Scene
	for _, n := range numbers {
		if n >= 1 && n <= len(p.Scenes) {
			out = append(out, p.Scenes[n-1])
		}
	}
	return out
}
