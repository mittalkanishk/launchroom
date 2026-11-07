package videocmd

import (
	"context"
	"strings"
	"testing"

	"github.com/mittalkanishk/launch-room/internal/cli"
	"github.com/mittalkanishk/launch-room/internal/videos"
)

type fakeMotionMaker struct {
	name  string
	shots []int
}

func (f *fakeMotionMaker) Motion(ctx context.Context, name string, shots []int) ([]videos.SceneResult, error) {
	f.name, f.shots = name, shots
	return []videos.SceneResult{
		{Shot: 1, File: "/videos/loop/renders/scenes/shot-01-motion.jpg", Status: "done"},
		{Shot: 3, Status: "failed", Message: "scene 3 isn't rendered yet"},
	}, nil
}

func TestMotionNamesEachStripAndHowToReadIt(t *testing.T) {
	env := &bufferEnv{}
	maker := &fakeMotionMaker{}
	cmd := NewMotion(env, maker)
	if code := cmd.Run(context.Background(), []string{"loop", "--shots", "1,3"}); code != cli.ExitError || maker.name != "loop" || len(maker.shots) != 2 {
		t.Fatalf("code = %d, asked for %q %v", code, maker.name, maker.shots)
	}
	got := env.stdout.String()
	for _, want := range []string{"scene 01: /videos/loop/renders/scenes/shot-01-motion.jpg", "scene 03: failed: scene 3 isn't rendered yet", "one row per sentence"} {
		if !strings.Contains(got, want) {
			t.Errorf("output lacks %q:\n%s", want, got)
		}
	}
	if code := cmd.Run(context.Background(), []string{"loop", "--look", "warm"}); code != cli.ExitUsage {
		t.Errorf("--look was accepted: code %d", code)
	}
}
