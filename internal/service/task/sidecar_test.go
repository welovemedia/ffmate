package task

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"github.com/welovemedia/ffmate/v2/internal/dto"
)

func TestReimportSidecarOnlyAppliesEditableFields(t *testing.T) {
	sidecar := filepath.Join(t.TempDir(), "side.json")
	err := os.WriteFile(sidecar, []byte(`{
		"uuid": "00000000-0000-0000-0000-000000000000",
		"status": "DONE_SUCCESSFUL",
		"error": "injected",
		"batch": "injected",
		"retries": 99,
		"preProcessing": {"error": "injected", "importSidecar": false},
		"name": "renamed-by-sidecar",
		"command": {"raw": "-i ${INPUT_FILE} new ${OUTPUT_FILE}"},
		"inputFile": {"raw": "/new/in.mp4"},
		"outputFile": {"raw": "/new/out.mp4"},
		"metadata": {"foo": "bar"},
		"priority": 7,
		"labels": ["gpu"]
	}`), 0600)
	assert.NoError(t, err)

	processor := &dto.PrePostProcessing{SidecarPath: &dto.RawResolved{Raw: sidecar, Resolved: sidecar}, ImportSidecar: true}
	task := &model.Task{
		UUID:          "original-uuid",
		Status:        dto.PreProcessing,
		Name:          "t",
		Retries:       1,
		PreProcessing: processor,
		Command:       &dto.RawResolved{Raw: "-i ${INPUT_FILE} ${OUTPUT_FILE}"},
		InputFile:     &dto.RawResolved{Raw: "/in.mp4"},
		OutputFile:    &dto.RawResolved{Raw: "/out.mp4"},
	}

	assert.NoError(t, (&Service{}).reimportSidecarIfNeeded(task, processor, "pre"))

	// protected fields stay untouched
	assert.Equal(t, "original-uuid", task.UUID)
	assert.Equal(t, dto.PreProcessing, task.Status)
	assert.Empty(t, task.Error)
	assert.Empty(t, task.Batch)
	assert.Equal(t, 1, task.Retries)
	assert.Same(t, processor, task.PreProcessing)
	assert.Empty(t, task.PreProcessing.Error)
	assert.True(t, task.PreProcessing.ImportSidecar)

	// editable fields are imported
	assert.Equal(t, "renamed-by-sidecar", task.Name)
	assert.Equal(t, "-i ${INPUT_FILE} new ${OUTPUT_FILE}", task.Command.Raw)
	assert.Equal(t, "/new/in.mp4", task.InputFile.Raw)
	assert.Equal(t, "/new/out.mp4", task.OutputFile.Raw)
	assert.Equal(t, "bar", (*task.Metadata)["foo"])
	assert.Equal(t, uint(7), task.Priority)
	assert.Equal(t, []model.Label{{Value: "gpu"}}, task.Labels)
}

func TestReimportSidecarKeepsAbsentFields(t *testing.T) {
	sidecar := filepath.Join(t.TempDir(), "side.json")
	assert.NoError(t, os.WriteFile(sidecar, []byte(`{"name": "only-name"}`), 0600))

	processor := &dto.PrePostProcessing{SidecarPath: &dto.RawResolved{Raw: sidecar, Resolved: sidecar}, ImportSidecar: true}
	task := &model.Task{
		Priority:  3,
		Command:   &dto.RawResolved{Raw: "cmd"},
		InputFile: &dto.RawResolved{Raw: "/in.mp4"},
		Labels:    []model.Label{{Value: "cpu"}},
	}

	assert.NoError(t, (&Service{}).reimportSidecarIfNeeded(task, processor, "pre"))
	assert.Equal(t, "only-name", task.Name)
	assert.Equal(t, uint(3), task.Priority)
	assert.Equal(t, "cmd", task.Command.Raw)
	assert.Equal(t, "/in.mp4", task.InputFile.Raw)
	assert.Equal(t, []model.Label{{Value: "cpu"}}, task.Labels)
}
