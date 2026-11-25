package task

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/cfg"
	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"github.com/welovemedia/ffmate/v2/internal/dto"
)

// writeArgScript creates a script that appends its arguments to log
func writeArgScript(t *testing.T, dir string) (script, log string) {
	t.Helper()
	script = filepath.Join(dir, "argv.sh")
	log = filepath.Join(dir, "argv.log")
	assert.NoError(t, os.WriteFile(script, []byte("#!/bin/sh\necho \"$@\" >> \""+log+"\"\n"), 0o755))
	return script, log
}

func TestOutputFileResolvedOnceForAllSteps(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell script")
	}
	cfg.Set("ffmate.ffmpeg", "ffmpeg")
	dir := t.TempDir()
	script, log := writeArgScript(t, dir)

	task := &model.Task{
		UUID:           "resolve-test",
		InputFile:      &dto.RawResolved{Raw: filepath.Join(dir, "in.mp4")},
		OutputFile:     &dto.RawResolved{Raw: filepath.Join(dir, "out-${UUID}.mp4")},
		Command:        &dto.RawResolved{Raw: "-i ${INPUT_FILE} ${OUTPUT_FILE}"},
		PreProcessing:  &dto.PrePostProcessing{ScriptPath: &dto.RawResolved{Raw: script + " PRE ${OUTPUT_FILE}"}},
		PostProcessing: &dto.PrePostProcessing{ScriptPath: &dto.RawResolved{Raw: script + " POST ${OUTPUT_FILE}"}},
	}
	s := &Service{}

	// same order as processNewTask
	s.resolveTaskFiles(task)
	assert.NoError(t, s.handleScriptExecution(task, task.PreProcessing, "pre"))
	s.prepareTaskFiles(task)
	assert.NoError(t, s.handleScriptExecution(task, task.PostProcessing, "post"))

	output := task.OutputFile.Resolved
	assert.NotContains(t, output, "${UUID}")
	assert.Contains(t, task.Command.Resolved, output, "ffmpeg command uses the same output file")

	data, err := os.ReadFile(log)
	assert.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	assert.Equal(t, []string{"PRE " + output, "POST " + output}, lines, "pre- and post-processing get the same output file")
}

func TestSidecarChangedOutputFileIsResolvedAgain(t *testing.T) {
	cfg.Set("ffmate.ffmpeg", "ffmpeg")
	dir := t.TempDir()
	sidecar := filepath.Join(dir, "side.json")
	assert.NoError(t, os.WriteFile(sidecar, []byte(`{"outputFile":{"raw":"`+filepath.Join(dir, "changed-${UUID}.mp4")+`"}}`), 0o600))

	processor := &dto.PrePostProcessing{SidecarPath: &dto.RawResolved{Raw: sidecar, Resolved: sidecar}, ImportSidecar: true}
	task := &model.Task{
		InputFile:  &dto.RawResolved{Raw: filepath.Join(dir, "in.mp4")},
		OutputFile: &dto.RawResolved{Raw: filepath.Join(dir, "out-${UUID}.mp4")},
		Command:    &dto.RawResolved{Raw: "-i ${INPUT_FILE} ${OUTPUT_FILE}"},
	}
	s := &Service{}

	s.resolveTaskFiles(task)
	input := task.InputFile.Resolved
	assert.NoError(t, s.reimportSidecarIfNeeded(task, processor, "pre"))
	s.prepareTaskFiles(task)

	assert.Equal(t, input, task.InputFile.Resolved, "unchanged input file keeps its resolved value")
	assert.Contains(t, task.OutputFile.Resolved, "changed-", "output file changed by the sidecar is resolved again")
	assert.NotContains(t, task.OutputFile.Resolved, "${UUID}")
}
