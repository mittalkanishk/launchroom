package renderer

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mittalkanishk/launch-room/internal/media"
	"github.com/mittalkanishk/launch-room/internal/run"
)

// Manim renders a Python scene, a `class Shot(Explainer)` written with the launchroom library, with
// Manim Community (the renderer container).
type Manim struct{}

func (Manim) SourceFile() string { return "scene.py" }

func (Manim) Check(source string) error {
	if !strings.Contains(source, "class Shot") {
		return errors.New("the scene must define class Shot(Explainer)")
	}
	return nil
}

func (Manim) Command(dir string) run.Cmd {
	return run.Cmd{Dir: dir, Name: "manim", Args: []string{
		"render", "scene.py", "Shot",
		"--resolution", fmt.Sprintf("%d,%d", media.Width, media.Height),
		"--frame_rate", fmt.Sprint(media.FrameRate),
		"--media_dir", filepath.Join(dir, "media"),
		"--verbosity", "WARNING", "--progress_bar", "none",
	}}
}

func (Manim) Outputs(dir string) (string, string, error) {
	videos, _ := filepath.Glob(filepath.Join(dir, "media", "videos", "scene", "*", "Shot.mp4"))
	if len(videos) == 0 {
		return "", "", errors.New("manim wrote no video")
	}
	return videos[0], strings.TrimSuffix(videos[0], ".mp4") + ".srt", nil
}

// Stage renders a JavaScript scene, `export default scene({ shots })`, with anime.js in headless
// Chromium (the stage container). Script is its renderer, docker/stage/runtime/render.mjs.
type Stage struct {
	Script string
}

func (Stage) SourceFile() string { return "scene.js" }

func (Stage) Check(source string) error {
	if !strings.Contains(source, "export default") {
		return errors.New("the scene must `export default scene({ shots: [...] })`")
	}
	return nil
}

func (s Stage) Command(dir string) run.Cmd {
	return run.Cmd{Dir: dir, Name: "node", Args: []string{
		s.Script, "--scene", "scene.js", "--out", filepath.Join(dir, "out"),
		"--width", fmt.Sprint(media.Width), "--height", fmt.Sprint(media.Height), "--fps", fmt.Sprint(media.FrameRate),
	}}
}

func (Stage) Outputs(dir string) (string, string, error) {
	video := filepath.Join(dir, "out", "scene.mp4")
	if _, err := os.Stat(video); err != nil {
		return "", "", errors.New("the stage wrote no video")
	}
	return video, filepath.Join(dir, "out", "scene.srt"), nil
}
