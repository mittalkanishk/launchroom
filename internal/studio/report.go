package studio

import (
	"encoding/json"
	"fmt"
	"math"
	"os"

	"github.com/mittalkanishk/launch-room/internal/media"
)

// Report is QC's measurement of a render against the rules.
type Report struct {
	Render       string             `json:"render"`
	Measurements media.Measurements `json:"measurements"`
	Checks       []Check            `json:"checks"`
}

// Check is one rule and whether the render passes it. A required check must pass before the video
// can be published; an advisory one is for the crew to weigh.
type Check struct {
	Name     string `json:"name"`
	Pass     bool   `json:"pass"`
	Required bool   `json:"required"`
	Detail   string `json:"detail"`
}

// Passed reports whether every required check passed.
func (r Report) Passed() bool {
	for _, c := range r.Checks {
		if c.Required && !c.Pass {
			return false
		}
	}
	return true
}

// Checks are the rules an explainer must meet before it is published: landscape 16:9 at full HD,
// one to ten minutes, and sound at platform loudness (-14 LUFS, peaks under -1 dBTP).
func Checks(m media.Measurements) []Check {
	return []Check{
		{
			Name:     "landscape 1920x1080",
			Required: true,
			Pass:     m.Width == media.Width && m.Height == media.Height,
			Detail:   fmt.Sprintf("%dx%d", m.Width, m.Height),
		},
		{
			Name:     "1 to 10 minutes",
			Required: true,
			Pass:     m.DurationSeconds >= 60 && m.DurationSeconds <= 600,
			Detail:   fmt.Sprintf("%.1fs", m.DurationSeconds),
		},
		{
			Name:     "loudness about -14 LUFS",
			Required: true,
			Pass:     m.HasAudio && math.Abs(m.LoudnessLUFS-(-14)) <= 2,
			Detail:   fmt.Sprintf("%.1f LUFS", m.LoudnessLUFS),
		},
		{
			Name:     "peaks under -1 dBTP",
			Required: true,
			Pass:     m.HasAudio && m.TruePeakDBTP <= -0.5,
			Detail:   fmt.Sprintf("%.1f dBTP", m.TruePeakDBTP),
		},
	}
}

func writeJSON(path string, v any) error {
	content, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(content, '\n'), 0o644)
}
