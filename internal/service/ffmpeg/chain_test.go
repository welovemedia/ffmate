package ffmpeg

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/cfg"
	"github.com/welovemedia/ffmate/v2/internal/database/model"
)

func TestSplitCommandChain(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"-i a.mp4 b.mp4", []string{"-i a.mp4 b.mp4"}},
		{"-i a.mp4 -f null - && ffmpeg -i a.mp4 b.mp4", []string{"-i a.mp4 -f null -", "ffmpeg -i a.mp4 b.mp4"}},
		{"-i a&&-i b", []string{"-i a", "-i b"}},
		{"-i 'a && b.mp4' out.mp4", []string{"-i 'a && b.mp4' out.mp4"}},
		{`-i "a && b.mp4" out.mp4`, []string{`-i "a && b.mp4" out.mp4`}},
		{`-i a\&\&b.mp4 out.mp4`, []string{`-i a\&\&b.mp4 out.mp4`}},
		{`-i 'it'"'"'s && x.mp4' o.mp4 && -i y o2.mp4`, []string{`-i 'it'"'"'s && x.mp4' o.mp4`, "-i y o2.mp4"}},
	}
	for _, c := range cases {
		got, err := splitCommandChain(c.in, false)
		assert.NoError(t, err, c.in)
		assert.Equal(t, c.want, got, c.in)
	}

	// windows: backslash is a path separator, not an escape
	got, err := splitCommandChain(`-i C:\in\a.mp4 && -i C:\in\b.mp4`, true)
	assert.NoError(t, err)
	assert.Equal(t, []string{`-i C:\in\a.mp4`, `-i C:\in\b.mp4`}, got)

	for _, bad := range []string{"-i 'a.mp4", "-i a.mp4 &&", "&& -i a.mp4", "-i a && && -i b"} {
		_, err := splitCommandChain(bad, false)
		assert.Error(t, err, bad)
	}
}

func requireFFmpeg(t *testing.T) string {
	t.Helper()
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not installed")
	}
	cfg.Set("ffmate.ffmpeg", path)
	return path
}

func runChain(t *testing.T, command string) error {
	t.Helper()
	return NewService().Execute(&ExecutionRequest{
		Ctx:        context.Background(),
		Task:       &model.Task{UUID: "chain-test"},
		Command:    command,
		UpdateFunc: func(float64, float64) {},
	})
}

func quote(v string) string { return "'" + strings.ReplaceAll(v, "'", `'"'"'`) + "'" }

func TestExecuteChainedCommands(t *testing.T) {
	ffmpegPath := requireFFmpeg(t)
	dir := t.TempDir()
	first, second, third := filepath.Join(dir, "first.mp4"), filepath.Join(dir, "second.mp4"), filepath.Join(dir, "third.mp4")
	src := "-y -f lavfi -i testsrc=d=0.2:s=64x64 "

	// documented form: following commands start with ${FFMPEG} (resolved and quoted), and the form without it
	command := src + quote(first) + " && " + quote(ffmpegPath) + " " + src + quote(second) + " && " + src + quote(third)
	assert.NoError(t, runChain(t, command))
	for _, f := range []string{first, second, third} {
		_, err := os.Stat(f)
		assert.NoError(t, err, "%s must be created", filepath.Base(f))
	}
}

func TestExecuteFileNameWithAmpersands(t *testing.T) {
	requireFFmpeg(t)
	dir := t.TempDir()
	input := filepath.Join(dir, "a && b.mp4")
	output := filepath.Join(dir, "out && more.mp4")
	assert.NoError(t, runChain(t, "-y -f lavfi -i testsrc=d=0.2:s=64x64 "+quote(input)))
	assert.NoError(t, runChain(t, "-y -i "+quote(input)+" "+quote(output)))
	_, err := os.Stat(output)
	assert.NoError(t, err)
}

func TestExecuteChainStopsOnFailure(t *testing.T) {
	requireFFmpeg(t)
	dir := t.TempDir()
	after := filepath.Join(dir, "after.mp4")
	err := runChain(t, "-y -i "+quote(filepath.Join(dir, "missing.mp4"))+" "+quote(filepath.Join(dir, "x.mp4"))+" && -y -f lavfi -i testsrc=d=0.2 "+quote(after))
	assert.Error(t, err)
	_, statErr := os.Stat(after)
	assert.True(t, os.IsNotExist(statErr), "the chain stops at the first failing command")
}
