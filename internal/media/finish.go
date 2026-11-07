package media

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// Looks are the colour grades the editor can choose, as ffmpeg filters. Each is applied the same
// way to the whole video, so the shots share one look.
var Looks = map[string]string{
	"neutral": "eq=contrast=1.03:saturation=1.05",
	"warm":    "colortemperature=temperature=5200,eq=contrast=1.05:saturation=1.12",
	"cool":    "colortemperature=temperature=8200,eq=contrast=1.05:saturation=1.0",
	"punchy":  "eq=contrast=1.15:saturation=1.3:gamma=0.98",
	"film":    "curves=preset=medium_contrast,eq=saturation=0.88,colorbalance=rs=0.04:bs=-0.03",
	"mono":    "hue=s=0,eq=contrast=1.12",
}

// LookNames is every look, sorted.
func LookNames() []string {
	names := make([]string, 0, len(Looks))
	for name := range Looks {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// -14 LUFS integrated with peaks under -1 dBTP: YouTube and the social platforms normalize to it.
const loudness = "loudnorm=I=-14:TP=-1:LRA=11"

// Finish grades the cut with one look and brings its sound to platform loudness, writing out.
func (t *Tools) Finish(ctx context.Context, in, out, look string) error {
	filter, ok := Looks[look]
	if !ok {
		return fmt.Errorf("unknown look %q; the looks are %s", look, strings.Join(LookNames(), ", "))
	}
	return t.runner.Run(ctx, ffmpeg("-i", in,
		"-vf", filter, "-af", loudness,
		"-c:v", "libx264", "-preset", "veryfast", "-crf", "20", "-pix_fmt", "yuv420p",
		"-c:a", "aac", "-b:a", "160k", "-ar", "48000",
		"-movflags", "+faststart", out))
}
