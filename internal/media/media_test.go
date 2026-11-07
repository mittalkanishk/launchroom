package media

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mittalkanishk/launch-room/internal/run"
)

// These tests run the real ffmpeg; they are skipped where it isn't installed.
func newTools(t *testing.T) *Tools {
	t.Helper()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg is not installed")
	}
	return NewTools(run.NewExec(io.Discard, io.Discard))
}

// testVideo is a 6-second 16:9 clip with a quiet tone.
func testVideo(t *testing.T, tools *Tools) string {
	t.Helper()
	out := filepath.Join(t.TempDir(), "in.mp4")
	err := tools.runner.Run(context.Background(), ffmpeg(
		"-f", "lavfi", "-i", "color=c=0x202020:s=1920x1080:d=2:r=25",
		"-f", "lavfi", "-i", "color=c=0xd0d0d0:s=1920x1080:d=4:r=25",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=6",
		"-filter_complex", "[0:v][1:v]concat=n=2:v=1:a=0[v];[2:a]volume=0.05[a]",
		"-map", "[v]", "-map", "[a]", "-c:v", "libx264", "-preset", "ultrafast", "-pix_fmt", "yuv420p",
		"-c:a", "aac", out))
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestFinishReachesSocialLoudness(t *testing.T) {
	tools := newTools(t)
	ctx := context.Background()
	in := testVideo(t, tools)

	before, err := tools.Measure(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	if before.Width != Width || before.Height != Height || !before.HasAudio || math.Abs(before.DurationSeconds-6) > 0.2 {
		t.Fatalf("test video measured %+v", before)
	}

	out := filepath.Join(t.TempDir(), "final.mp4")
	if err := tools.Finish(ctx, in, out, "warm"); err != nil {
		t.Fatal(err)
	}
	after, err := tools.Measure(ctx, out)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(after.LoudnessLUFS-(-14)) > 1.5 {
		t.Errorf("loudness %.1f LUFS after finishing (was %.1f), want about -14", after.LoudnessLUFS, before.LoudnessLUFS)
	}
	if err := tools.Finish(ctx, in, out, "sepia-dream"); err == nil {
		t.Error("an unknown look was accepted")
	}
}

func TestKeyframes(t *testing.T) {
	tools := newTools(t)
	frames, err := tools.Keyframes(context.Background(), testVideo(t, tools), t.TempDir(), 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(frames) < 2 {
		t.Errorf("keyframes = %v, want one every 3s of 6s", frames)
	}
}

func TestJoinScenesWritesOffsetSubtitlesBurnedInOrNot(t *testing.T) {
	tools := newTools(t)
	ctx := context.Background()
	if out, _ := exec.Command("ffmpeg", "-hide_banner", "-filters").Output(); !strings.Contains(string(out), " subtitles ") {
		t.Skip("this ffmpeg has no subtitles filter (libass)")
	}
	dir := t.TempDir()

	var scenes, subs []string
	for i, color := range []string{"0x202040", "0x204020"} {
		scene := filepath.Join(dir, fmt.Sprintf("scene-%d.mp4", i))
		err := tools.runner.Run(ctx, ffmpeg("-f", "lavfi", "-i", "color=c="+color+":s=960x540:d=2:r=15",
			"-f", "lavfi", "-i", "sine=frequency=330:duration=2", "-c:v", "libx264", "-preset", "ultrafast",
			"-pix_fmt", "yuv420p", "-c:a", "aac", "-shortest", scene))
		if err != nil {
			t.Fatal(err)
		}
		sub := filepath.Join(dir, fmt.Sprintf("scene-%d.srt", i))
		os.WriteFile(sub, []byte(fmt.Sprintf("1\n00:00:00,100 --> 00:00:01,800\nScene %d says hello\n\n", i)), 0o644)
		scenes, subs = append(scenes, scene), append(subs, sub)
	}

	for _, burn := range []bool{true, false} {
		sub := filepath.Join(dir, fmt.Sprintf("burn-%v", burn))
		os.MkdirAll(sub, 0o755)
		out := filepath.Join(sub, "cut.mp4")
		if err := tools.JoinScenes(ctx, scenes, subs, burn, out); err != nil {
			t.Fatalf("burn=%v: %v", burn, err)
		}
		m, err := tools.Measure(ctx, out)
		if err != nil {
			t.Fatal(err)
		}
		if m.Width != Width || m.Height != Height || math.Abs(m.DurationSeconds-4) > 0.3 || !m.HasAudio {
			t.Errorf("burn=%v: joined video measured %+v", burn, m)
		}
		// The joined SRT is written either way, so the words can be uploaded as captions.
		srt, _ := os.ReadFile(filepath.Join(sub, "cut.srt"))
		if !strings.Contains(string(srt), "00:00:02,100 --> 00:00:03,800\nScene 1 says hello") {
			t.Errorf("burn=%v: combined subtitles = %q, want scene 1's cue offset by 2s", burn, srt)
		}
	}
}

func TestSheetHasAFramePerSentence(t *testing.T) {
	tools := newTools(t)
	dir := t.TempDir()
	srt := filepath.Join(dir, "scene.srt")
	os.WriteFile(srt, []byte("1\n00:00:00,500 --> 00:00:01,500\nDark.\n\n2\n00:00:03,000 --> 00:00:05,000\nLight.\n"), 0o644)
	out := filepath.Join(dir, "sheet.jpg")
	if err := tools.Sheet(context.Background(), testVideo(t, tools), srt, out); err != nil {
		t.Fatal(err)
	}
	var m Measurements
	if err := tools.probe(context.Background(), out, &m); err != nil {
		t.Fatal(err)
	}
	// Two frames side by side at 480 wide, with the tile's 6px margins and padding.
	if m.Width != 2*480+3*6 || m.Height != 270+2*6 {
		t.Errorf("sheet is %dx%d, want two 480x270 frames tiled", m.Width, m.Height)
	}
}

func TestMotionHasARowOfEightFramesPerSentence(t *testing.T) {
	tools := newTools(t)
	dir := t.TempDir()
	srt := filepath.Join(dir, "scene.srt")
	os.WriteFile(srt, []byte("1\n00:00:00,500 --> 00:00:01,500\nDark.\n\n2\n00:00:03,000 --> 00:00:05,000\nLight.\n"), 0o644)
	out := filepath.Join(dir, "motion.jpg")
	if err := tools.Motion(context.Background(), testVideo(t, tools), srt, out); err != nil {
		t.Fatal(err)
	}
	var m Measurements
	if err := tools.probe(context.Background(), out, &m); err != nil {
		t.Fatal(err)
	}
	// Two rows (one per sentence) of eight 300x168 frames, with the tile's 6px margins and padding.
	if m.Width != 8*300+9*6 || m.Height != 2*168+3*6 {
		t.Errorf("motion strip is %dx%d, want 2 rows of 8 frames at 300x168", m.Width, m.Height)
	}
}
