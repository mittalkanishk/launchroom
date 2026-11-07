package studio

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mittalkanishk/launch-room/internal/media"
	"github.com/mittalkanishk/launch-room/internal/names"
)

const maxScenes = 60

// check refuses a request that can't run.
func (s *Service) check(req Request) error {
	if err := names.Check(req.Video); err != nil {
		return err
	}
	switch req.Kind {
	case KindScene:
		if req.Shot < 1 || req.Shot > maxScenes {
			return fmt.Errorf("a scene is numbered 1 to %d", maxScenes)
		}
		if _, ok := s.renderers[req.Engine]; !ok {
			return fmt.Errorf("unknown engine %q; a scene is rendered by manim or stage", req.Engine)
		}
		if req.Source == "" {
			return errors.New("a scene needs its source")
		}
	case KindCut:
		if len(req.Shots) == 0 {
			return errors.New("a cut needs its scenes")
		}
		for _, shot := range req.Shots {
			if _, err := os.Stat(s.sceneFile(req.Video, shot)); err != nil {
				return fmt.Errorf("scene %d isn't rendered yet", shot)
			}
		}
	case KindFinish:
		if _, ok := media.Looks[req.Look]; !ok {
			return fmt.Errorf("unknown look %q; the looks are %v", req.Look, media.LookNames())
		}
		if _, err := s.File(req.Video, "cut.mp4"); err != nil {
			return errors.New("there is no cut to finish yet")
		}
	case KindQC:
		if _, err := s.latestRender(req.Video); err != nil {
			return err
		}
	case KindMotion:
		if req.Shot < 1 || req.Shot > maxScenes {
			return fmt.Errorf("a scene is numbered 1 to %d", maxScenes)
		}
		if _, err := os.Stat(s.sceneFile(req.Video, req.Shot)); err != nil {
			return fmt.Errorf("scene %d isn't rendered yet", req.Shot)
		}
	default:
		return fmt.Errorf("unknown kind %q", req.Kind)
	}
	return nil
}

// step runs the request's kind and returns the files it wrote.
func (s *Service) step(ctx context.Context, req Request) ([]string, error) {
	dir := s.videoDir(req.Video)
	for _, sub := range []string{"scenes", "qc"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0o755); err != nil {
			return nil, err
		}
	}
	switch req.Kind {
	case KindScene:
		return s.scene(ctx, req)
	case KindCut:
		return s.cut(ctx, req)
	case KindFinish:
		return []string{"final.mp4"}, s.post.Finish(ctx, filepath.Join(dir, "cut.mp4"), filepath.Join(dir, "final.mp4"), req.Look)
	case KindQC:
		return s.qc(ctx, req.Video)
	case KindMotion:
		return s.motion(ctx, req)
	}
	return nil, fmt.Errorf("unknown kind %q", req.Kind)
}

// motion saves a rendered scene's motion strip: a row of frames per sentence, for judging how
// things move between the sheet's keyframes.
func (s *Service) motion(ctx context.Context, req Request) ([]string, error) {
	dir := s.videoDir(req.Video)
	strip := sceneName(req.Shot, "-motion.jpg")
	err := s.post.Motion(ctx, filepath.Join(dir, sceneName(req.Shot, ".mp4")), filepath.Join(dir, sceneName(req.Shot, ".srt")), filepath.Join(dir, strip))
	return []string{strip}, err
}

// scene renders one scene and saves its video, its subtitles, and its sheet: a frame per
// sentence, for the director and reviewer to look at.
func (s *Service) scene(ctx context.Context, req Request) ([]string, error) {
	result, err := s.renderers[req.Engine].Render(ctx, req.Source)
	if err != nil {
		return nil, err
	}
	video := sceneName(req.Shot, ".mp4")
	subtitles := sceneName(req.Shot, ".srt")
	sheet := sceneName(req.Shot, ".jpg")
	dir := s.videoDir(req.Video)
	if err := os.WriteFile(filepath.Join(dir, video), result.Video, 0o644); err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(dir, subtitles), []byte(result.Subtitles), 0o644); err != nil {
		return nil, err
	}
	if err := s.post.Sheet(ctx, filepath.Join(dir, video), filepath.Join(dir, subtitles), filepath.Join(dir, sheet)); err != nil {
		return nil, err
	}
	return []string{video, subtitles, sheet}, nil
}

// cut joins the scenes, each as last rendered. Their subtitles are burned in unless the request
// says not to; either way the joined SRT is written beside the cut.
func (s *Service) cut(ctx context.Context, req Request) ([]string, error) {
	var scenes, subtitles []string
	for _, shot := range req.Shots {
		scene := s.sceneFile(req.Video, shot)
		scenes = append(scenes, scene)
		subtitles = append(subtitles, strings.TrimSuffix(scene, ".mp4")+".srt")
	}
	return []string{"cut.mp4", "cut.srt"}, s.post.JoinScenes(ctx, scenes, subtitles, !req.NoSubtitles, filepath.Join(s.videoDir(req.Video), "cut.mp4"))
}

// qc measures the latest render against the rules and saves a keyframe every three seconds.
func (s *Service) qc(ctx context.Context, video string) ([]string, error) {
	render, err := s.latestRender(video)
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(s.videoDir(video), "qc")
	old, _ := filepath.Glob(filepath.Join(dir, "keyframe-*.jpg"))
	for _, f := range old {
		os.Remove(f)
	}
	m, err := s.post.Measure(ctx, filepath.Join(s.videoDir(video), render))
	if err != nil {
		return nil, err
	}
	frames, err := s.post.Keyframes(ctx, filepath.Join(s.videoDir(video), render), dir, 3)
	if err != nil {
		return nil, err
	}
	report := Report{Render: render, Measurements: m, Checks: Checks(m)}
	if err := writeJSON(filepath.Join(dir, "report.json"), report); err != nil {
		return nil, err
	}
	files := []string{"qc/report.json"}
	for _, f := range frames {
		files = append(files, filepath.Join("qc", filepath.Base(f)))
	}
	return files, nil
}

// latestRender is the most finished render of the video: final, else cut.
func (s *Service) latestRender(video string) (string, error) {
	for _, rel := range []string{"final.mp4", "cut.mp4"} {
		if _, err := s.File(video, rel); err == nil {
			return rel, nil
		}
	}
	return "", errors.New("there is no render to check yet")
}

func (s *Service) sceneFile(video string, shot int) string {
	return filepath.Join(s.videoDir(video), sceneName(shot, ".mp4"))
}

func sceneName(shot int, ext string) string {
	return filepath.Join("scenes", fmt.Sprintf("shot-%02d%s", shot, ext))
}
