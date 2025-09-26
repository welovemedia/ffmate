package task

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/mattn/go-shellwords"
	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"github.com/welovemedia/ffmate/v2/internal/debug"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"github.com/welovemedia/ffmate/v2/internal/metrics"
	"github.com/welovemedia/ffmate/v2/internal/service/ffmpeg"
)

func (s *Service) runPreProcessing(task *model.Task) error {
	if err := s.prePostProcessTask(task, task.PreProcessing, "pre"); err != nil {
		return fmt.Errorf("PreProcessing failed: %v", err)
	}
	return nil
}

func (s *Service) prepareTaskFiles(task *model.Task) {
	task.InputFile.Resolved = s.wildcardReplacer(task.InputFile.Raw, task.InputFile.Raw, task.OutputFile.Raw, task.Source, task.Metadata)
	task.OutputFile.Resolved = s.wildcardReplacer(task.OutputFile.Raw, task.InputFile.Raw, task.OutputFile.Raw, task.Source, task.Metadata)
	task.Command.Resolved = s.wildcardReplacer(task.Command.Raw, task.InputFile.Resolved, task.OutputFile.Resolved, task.Source, task.Metadata)

	task.Status = dto.Running
	if _, err := s.Update(task); err != nil {
		debug.Log.Error("failed to save task (uuid: %s)", task.UUID)
	}
}

func (s *Service) createOutputDirectory(task *model.Task) error {
	if err := os.MkdirAll(filepath.Dir(task.OutputFile.Resolved), 0755); err != nil {
		return fmt.Errorf("failed to create non-existing output directory: %v", err)
	}
	return nil
}

func (s *Service) executeFFmpeg(task *model.Task) error {
	debug.Task.Debug("starting ffmpeg process (uuid: %s)", task.UUID)

	ctxAny, _ := taskQueue.Load(task.UUID)
	ctx := ctxAny.(context.Context)

	err := s.ffmpegService.Execute(&ffmpeg.ExecutionRequest{
		Task:    task,
		Command: task.Command.Resolved,
		Ctx:     ctx,
		UpdateFunc: func(progress, remaining float64) {
			task.Progress = progress
			task.Remaining = remaining
			if _, err := s.Update(task); err != nil {
				debug.Log.Error("failed to save task (uuid: %s)", task.UUID)
			}
		},
	})

	task.Progress = 100
	task.Remaining = -1

	if err != nil {
		debug.Task.Debug("finished processing with error (uuid: %s): %v", task.UUID, err)

		if cause := context.Cause(ctx); cause != nil {
			s.cancelTask(task, cause)
			return cause
		}

		s.failTask(task, err)
		return err
	}

	debug.Task.Debug("finished processing (uuid: %s)", task.UUID)
	return nil
}

func (s *Service) runPostProcessing(task *model.Task) error {
	if err := s.prePostProcessTask(task, task.PostProcessing, "post"); err != nil {
		return fmt.Errorf("PostProcessing failed: %v", err)
	}
	return nil
}

func (s *Service) finalizeTask(task *model.Task) {
	task.FinishedAt = time.Now().UnixMilli()
	task.Status = dto.DoneSuccessful
	if _, err := s.Update(task); err != nil {
		debug.Log.Error("failed to save task (uuid: %s)", task.UUID)
	}
	debug.Task.Info("task successful (uuid: %s)", task.UUID)
}

func (s *Service) prePostProcessTask(task *model.Task, processor *dto.PrePostProcessing, processorType string) error {
	if processor == nil || (processor.SidecarPath == nil && processor.ScriptPath == nil) {
		return nil
	}

	s.trackProcessingMetrics(processor, processorType)
	s.initializeProcessing(task, processor, processorType)

	if err := s.handleSidecar(task, processor, processorType); err != nil {
		return err
	}

	if err := s.handleScriptExecution(task, processor, processorType); err != nil {
		return err
	}

	if err := s.reimportSidecarIfNeeded(task, processor, processorType); err != nil {
		return err
	}

	return s.finalizeProcessing(processor, processorType, task)
}

func (s *Service) trackProcessingMetrics(processor *dto.PrePostProcessing, processorType string) {
	sidecarEmpty := strconv.FormatBool(processor.SidecarPath != nil && processor.SidecarPath.Raw == "")
	scriptEmpty := strconv.FormatBool(processor.ScriptPath != nil && processor.ScriptPath.Raw == "")

	metricName := "task.preProcessing"
	if processorType == "post" {
		metricName = "task.postProcessing"
	}
	metrics.GaugeVec(metricName).WithLabelValues(sidecarEmpty, scriptEmpty).Inc()
}

func (s *Service) initializeProcessing(task *model.Task, processor *dto.PrePostProcessing, processorType string) {
	debug.Task.Debug("starting %sProcessing (uuid: %s)", processorType, task.UUID)
	processor.StartedAt = time.Now().UnixMilli()

	if processorType == "pre" {
		task.Status = dto.PreProcessing
	} else {
		task.Status = dto.PostProcessing
	}

	if _, err := s.Update(task); err != nil {
		debug.Log.Error("failed to save task (uuid: %s)", task.UUID)
	}
}

func (s *Service) handleSidecar(task *model.Task, processor *dto.PrePostProcessing, processorType string) error {
	if processor.SidecarPath == nil || processor.SidecarPath.Raw == "" {
		return nil
	}

	// Resolve path and save
	if processorType == "pre" {
		processor.SidecarPath.Resolved = s.wildcardReplacer(processor.SidecarPath.Raw, task.InputFile.Raw, task.OutputFile.Raw, task.Source, task.Metadata)
	} else {
		processor.SidecarPath.Resolved = s.wildcardReplacer(processor.SidecarPath.Raw, task.InputFile.Resolved, task.OutputFile.Resolved, task.Source, task.Metadata)
	}
	if _, err := s.Update(task); err != nil {
		debug.Log.Error("failed to save task (uuid: %s)", task.UUID)
	}

	// Write file
	data, err := json.Marshal(task.ToDTO())
	if err != nil {
		debug.Log.Error("failed to marshal task to write sidecar file: %v", err)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(processor.SidecarPath.Resolved), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(processor.SidecarPath.Resolved, data, 0644); err != nil {
		processor.Error = fmt.Errorf("failed to write sidecar: %v", err).Error()
		debug.Log.Error("failed to write sidecar file: %v", err)
	}
	return nil
}

func (s *Service) handleScriptExecution(task *model.Task, processor *dto.PrePostProcessing, processorType string) error {
	if processor.Error != "" || processor.ScriptPath == nil || processor.ScriptPath.Raw == "" {
		return nil
	}

	if processorType == "pre" {
		processor.ScriptPath.Resolved = s.wildcardReplacer(processor.ScriptPath.Raw, task.InputFile.Raw, task.OutputFile.Raw, task.Source, task.Metadata)
	} else {
		processor.ScriptPath.Resolved = s.wildcardReplacer(processor.ScriptPath.Raw, task.InputFile.Resolved, task.OutputFile.Resolved, task.Source, task.Metadata)
	}
	if _, err := s.Update(task); err != nil {
		debug.Log.Error("failed to save task (uuid: %s)", task.UUID)
	}

	args, err := shellwords.NewParser().Parse(processor.ScriptPath.Resolved)
	if err != nil {
		processor.Error = err.Error()
		debug.Task.Debug("failed to parse %sProcessing script (uuid: %s): %v", processorType, task.UUID, err)
		return nil
	}

	cmd := exec.Command(args[0], args[1:]...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil || cmd.Wait() != nil {
		processor.Error = fmt.Sprintf("%s (exit code: %d)", stderr.String(), cmd.ProcessState.ExitCode())
		debug.Task.Debug("script failed (uuid: %s): stderr: %s", task.UUID, stderr.String())
	}
	return nil
}

func (s *Service) reimportSidecarIfNeeded(task *model.Task, processor *dto.PrePostProcessing, processorType string) error {
	if processorType != "pre" || processor.SidecarPath == nil || processor.SidecarPath.Raw == "" || !processor.ImportSidecar {
		return nil
	}
	data, err := os.ReadFile(processor.SidecarPath.Resolved)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, task); err != nil {
		return err
	}
	debug.Task.Debug("re-imported sidecar file (uuid: %s)", task.UUID)
	return nil
}

func (s *Service) finalizeProcessing(processor *dto.PrePostProcessing, processorType string, task *model.Task) error {
	processor.FinishedAt = time.Now().UnixMilli()
	if processor.Error != "" {
		debug.Task.Info("finished %sProcessing with error (uuid: %s)", processorType, task.UUID)
		return errors.New(processor.Error)
	}
	debug.Task.Info("finished %sProcessing (uuid: %s)", processorType, task.UUID)
	return nil
}
