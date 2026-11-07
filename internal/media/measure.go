package media

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mittalkanishk/launch-room/internal/run"
)

// Measurements are the facts QC checks against the platform rules.
type Measurements struct {
	DurationSeconds float64 `json:"duration_seconds"`
	Width           int     `json:"width"`
	Height          int     `json:"height"`
	HasAudio        bool    `json:"has_audio"`
	LoudnessLUFS    float64 `json:"loudness_lufs"`
	TruePeakDBTP    float64 `json:"true_peak_dbtp"`
}

// Measure reads the video's size, length and loudness.
func (t *Tools) Measure(ctx context.Context, in string) (Measurements, error) {
	var m Measurements
	if err := t.probe(ctx, in, &m); err != nil {
		return m, err
	}
	if m.HasAudio {
		if err := t.loudness(ctx, in, &m); err != nil {
			return m, err
		}
	}
	return m, nil
}

func (t *Tools) probe(ctx context.Context, in string, m *Measurements) error {
	out, err := t.runner.Output(ctx, run.Cmd{Name: "ffprobe", Args: []string{
		"-v", "error", "-show_entries", "stream=codec_type,width,height:format=duration", "-of", "json", in}})
	if err != nil {
		return err
	}
	var probe struct {
		Streams []struct {
			CodecType string `json:"codec_type"`
			Width     int    `json:"width"`
			Height    int    `json:"height"`
		} `json:"streams"`
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}
	if err := json.Unmarshal([]byte(out), &probe); err != nil {
		return fmt.Errorf("ffprobe %s: %w", in, err)
	}
	for _, s := range probe.Streams {
		switch s.CodecType {
		case "video":
			m.Width, m.Height = s.Width, s.Height
		case "audio":
			m.HasAudio = true
		}
	}
	m.DurationSeconds, _ = strconv.ParseFloat(probe.Format.Duration, 64)
	return nil
}

func (t *Tools) loudness(ctx context.Context, in string, m *Measurements) error {
	log, err := t.runner.Log(ctx, ffmpeg("-i", in, "-af", loudness+":print_format=json", "-vn", "-f", "null", "-"))
	if err != nil {
		return err
	}
	start := strings.LastIndex(log, "{")
	end := strings.LastIndex(log, "}")
	if start < 0 || end < start {
		return fmt.Errorf("%s: no loudness report", in)
	}
	var report struct {
		InputI  string `json:"input_i"`
		InputTP string `json:"input_tp"`
	}
	if err := json.Unmarshal([]byte(log[start:end+1]), &report); err != nil {
		return fmt.Errorf("%s: loudness report: %w", in, err)
	}
	m.LoudnessLUFS, _ = strconv.ParseFloat(report.InputI, 64)
	m.TruePeakDBTP, _ = strconv.ParseFloat(report.InputTP, 64)
	return nil
}

// Keyframes saves a small frame every few seconds into dir, for a reviewer who can look at
// pictures but not watch video. It returns the files written.
func (t *Tools) Keyframes(ctx context.Context, in, dir string, everySeconds int) ([]string, error) {
	pattern := filepath.Join(dir, "keyframe-%02d.jpg")
	if err := t.runner.Run(ctx, ffmpeg("-i", in,
		"-vf", fmt.Sprintf("fps=1/%d,scale=540:-2", everySeconds), "-q:v", "4", pattern)); err != nil {
		return nil, err
	}
	return filepath.Glob(filepath.Join(dir, "keyframe-*.jpg"))
}
