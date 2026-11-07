package media

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// JoinScenes cuts rendered explainer scenes together in order, at full HD. The narration's
// subtitles, each scene's offset by the scenes before it, are always written beside the cut as one
// SRT (for upload as captions); with burn they are also drawn into the picture.
func (t *Tools) JoinScenes(ctx context.Context, scenes, subtitles []string, burn bool, out string) error {
	if len(scenes) == 0 {
		return fmt.Errorf("no scenes to join")
	}
	srt := strings.TrimSuffix(out, filepath.Ext(out)) + ".srt"
	if err := t.combineSubtitles(ctx, scenes, subtitles, srt); err != nil {
		return err
	}

	var args []string
	var filter strings.Builder
	for i, scene := range scenes {
		args = append(args, "-i", scene)
		fmt.Fprintf(&filter, "[%d:v]scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2,setsar=1,fps=%d[v%d];",
			i, Width, Height, Width, Height, FrameRate, i)
		fmt.Fprintf(&filter, "[%d:a]aresample=48000,aformat=channel_layouts=stereo[a%d];", i, i)
	}
	for i := range scenes {
		fmt.Fprintf(&filter, "[v%d][a%d]", i, i)
	}
	if burn {
		style := "FontName=DejaVu Sans,FontSize=13,PrimaryColour=&H00EEEEEE,OutlineColour=&H00111111,BorderStyle=1,Outline=1.2,Shadow=0,MarginV=18"
		fmt.Fprintf(&filter, "concat=n=%d:v=1:a=1[joined][a];[joined]subtitles='%s':force_style='%s'[v]", len(scenes), escapeFilterPath(srt), style)
	} else {
		fmt.Fprintf(&filter, "concat=n=%d:v=1:a=1[v][a]", len(scenes))
	}

	args = append(args, "-filter_complex", filter.String(), "-map", "[v]", "-map", "[a]",
		"-c:v", "libx264", "-preset", "veryfast", "-crf", "20", "-pix_fmt", "yuv420p",
		"-c:a", "aac", "-b:a", "160k", "-movflags", "+faststart", out)
	return t.runner.Run(ctx, ffmpeg(args...))
}

// combineSubtitles writes one SRT for the joined video from each scene's SRT.
func (t *Tools) combineSubtitles(ctx context.Context, scenes, subtitles []string, out string) error {
	var b strings.Builder
	var offset time.Duration
	n := 1
	for i, scene := range scenes {
		if i < len(subtitles) && subtitles[i] != "" {
			cues, err := readSRT(subtitles[i])
			if err != nil {
				return err
			}
			for _, c := range cues {
				fmt.Fprintf(&b, "%d\n%s --> %s\n%s\n\n", n, srtTime(c.start+offset), srtTime(c.end+offset), c.text)
				n++
			}
		}
		var m Measurements
		if err := t.probe(ctx, scene, &m); err != nil {
			return err
		}
		offset += time.Duration(m.DurationSeconds * float64(time.Second))
	}
	return os.WriteFile(out, []byte(b.String()), 0o644)
}

type cue struct {
	start, end time.Duration
	text       string
}

func readSRT(path string) ([]cue, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var cues []cue
	var cur *cue
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case strings.Contains(line, "-->"):
			from, to, _ := strings.Cut(line, "-->")
			cur = &cue{start: parseSRTTime(from), end: parseSRTTime(to)}
		case line == "" && cur != nil:
			cues = append(cues, *cur)
			cur = nil
		case cur != nil:
			cur.text = strings.TrimSpace(cur.text + "\n" + line)
		}
	}
	if cur != nil {
		cues = append(cues, *cur)
	}
	return cues, scanner.Err()
}

// parseSRTTime reads "00:01:02,345".
func parseSRTTime(s string) time.Duration {
	s = strings.ReplaceAll(strings.TrimSpace(s), ",", ".")
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return 0
	}
	h, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	sec, _ := strconv.ParseFloat(parts[2], 64)
	return time.Duration(h)*time.Hour + time.Duration(m)*time.Minute + time.Duration(sec*float64(time.Second))
}

func srtTime(d time.Duration) string {
	ms := d.Milliseconds()
	return fmt.Sprintf("%02d:%02d:%02d,%03d", ms/3600000, ms/60000%60, ms/1000%60, ms%1000)
}

// escapeFilterPath makes a path safe inside a quoted ffmpeg filter option.
func escapeFilterPath(p string) string {
	return strings.NewReplacer(`\`, `\\`, `'`, `\'`, `:`, `\:`).Replace(p)
}
