package task

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/cfg"
	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"github.com/welovemedia/ffmate/v2/internal/dto"
)

func runScript(t *testing.T, script string) string {
	t.Helper()
	cfg.Set("ffmate.ffmpeg", "ffmpeg") // used by the wildcard replacer
	task := &model.Task{
		UUID:       "script-test",
		InputFile:  &dto.RawResolved{Raw: "/in.mp4"},
		OutputFile: &dto.RawResolved{Raw: "/out.mp4"},
	}
	processor := &dto.PrePostProcessing{ScriptPath: &dto.RawResolved{Raw: script}}
	assert.NoError(t, (&Service{}).handleScriptExecution(task, processor, "pre"))
	return processor.Error
}

func TestScriptNotFoundReportsReason(t *testing.T) {
	msg := runScript(t, filepath.Join(t.TempDir(), "nope", "x.sh"))
	assert.Contains(t, msg, "failed to start script")
	assert.Contains(t, msg, "no such file or directory")
}

func TestScriptNotExecutableReportsReason(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("file modes are not applicable on windows")
	}
	script := filepath.Join(t.TempDir(), "noexec.sh")
	assert.NoError(t, os.WriteFile(script, []byte("#!/bin/sh\nexit 0\n"), 0o644))
	msg := runScript(t, script)
	assert.Contains(t, msg, "failed to start script")
	assert.Contains(t, msg, "permission denied")
}

func TestScriptFailureKeepsStderrAndExitCode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell script")
	}
	script := filepath.Join(t.TempDir(), "fail.sh")
	assert.NoError(t, os.WriteFile(script, []byte("#!/bin/sh\necho boom >&2\nexit 3\n"), 0o755))
	msg := runScript(t, script)
	assert.Contains(t, msg, "boom")
	assert.Contains(t, msg, "(exit code: 3)")
}

func TestScriptFailureWithoutStderrUsesProcessError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell script")
	}
	script := filepath.Join(t.TempDir(), "silent.sh")
	assert.NoError(t, os.WriteFile(script, []byte("#!/bin/sh\nexit 4\n"), 0o755))
	msg := runScript(t, script)
	assert.Equal(t, "exit status 4 (exit code: 4)", msg)
}

func TestScriptSuccessHasNoError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell script")
	}
	script := filepath.Join(t.TempDir(), "ok.sh")
	assert.NoError(t, os.WriteFile(script, []byte("#!/bin/sh\nexit 0\n"), 0o755))
	assert.Empty(t, runScript(t, script))
}
