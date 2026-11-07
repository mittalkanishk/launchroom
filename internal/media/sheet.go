package media

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// sheetColumns is how many frames sit side by side on a scene's sheet.
const sheetColumns = 4

// Sheet saves one picture of a rendered scene: a frame from the middle of every spoken sentence
// (the subtitles say when each is said), tiled four across, for the director and reviewer, who can
// look at pictures but not watch video. A scene with no subtitles gets its first frame.
func (t *Tools) Sheet(ctx context.Context, in, subtitles, out string) error {
	cues, err := readSRT(subtitles)
	if err != nil {
		return err
	}
	var times []float64
	for _, c := range cues {
		times = append(times, (c.start+c.end).Seconds()/2)
	}
	if len(times) == 0 {
		times = []float64{0}
	}
	cols := min(sheetColumns, len(times))
	rows := (len(times) + sheetColumns - 1) / sheetColumns
	filter := fmt.Sprintf("select='%s',scale=480:-2,tile=%dx%d:padding=6:margin=6", selectAt(times), cols, rows)
	return t.runner.Run(ctx, ffmpeg("-i", in, "-vf", filter, "-fps_mode", "vfr", "-frames:v", "1", "-update", "1", "-q:v", "4", out))
}

// motionFrames is how many frames of each sentence a motion strip shows, left to right.
const motionFrames = 8

// Motion saves one picture of how a rendered scene moves: for every spoken sentence, a row of
// frames spread evenly across the time it is said, so a reviewer who can only look at pictures
// can judge a path (a note through a slot, a dot round a ring) and not just its endpoints, which
// is all the Sheet shows. A scene with no subtitles gets one row across the whole scene.
func (t *Tools) Motion(ctx context.Context, in, subtitles, out string) error {
	cues, err := readSRT(subtitles)
	if err != nil {
		return err
	}
	if len(cues) == 0 {
		var m Measurements
		if err := t.probe(ctx, in, &m); err != nil {
			return err
		}
		cues = []cue{{end: time.Duration(m.DurationSeconds * float64(time.Second))}}
	}
	var times []float64
	rows := 0
	for _, c := range cues {
		if c.end <= c.start {
			continue
		}
		rows++
		step := (c.end - c.start).Seconds() / motionFrames
		for i := range motionFrames {
			times = append(times, c.start.Seconds()+step*(float64(i)+0.5))
		}
	}
	if rows == 0 {
		return fmt.Errorf("no sentence in %s lasts long enough to show its motion", subtitles)
	}
	filter := fmt.Sprintf("select='%s',scale=300:-2,tile=%dx%d:padding=6:margin=6", selectAt(times), motionFrames, rows)
	return t.runner.Run(ctx, ffmpeg("-i", in, "-vf", filter, "-fps_mode", "vfr", "-frames:v", "1", "-update", "1", "-q:v", "4", out))
}

// selectAt is a select expression that passes exactly one frame at or after each time, in order:
// the first frame whose time has reached it and whose previous pick hadn't.
func selectAt(times []float64) string {
	terms := make([]string, len(times))
	for i, at := range times {
		terms[i] = fmt.Sprintf("gte(t\\,%.3f)*(isnan(prev_selected_t)+lt(prev_selected_t\\,%.3f))", at, at)
	}
	return strings.Join(terms, "+")
}
