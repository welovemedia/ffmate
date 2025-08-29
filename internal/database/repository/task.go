package repository

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/welovemedia/ffmate/internal/database/model"
	"github.com/welovemedia/ffmate/internal/dto"
	"gorm.io/gorm"
)

type statusCount struct {
	Status string
	Count  int
}

type Task struct {
	DB *gorm.DB
}

func (t *Task) Setup() {
	err := t.DB.AutoMigrate(&model.Task{})
	if err != nil {
		fmt.Printf("failed to initialize database: %v", err)
	}
}

func (m *Task) CountAllStatus(session string) (queued, running, doneSuccessful, doneError, doneCanceled int, err error) {
	var counts []statusCount

	if session != "" {
		m.DB.Model(&model.Task{}).
			Select("status, COUNT(*) as count").
			Group("status").
			Where("session = ?", session).
			Find(&counts)
	} else {
		m.DB.Model(&model.Task{}).
			Select("status, COUNT(*) as count").
			Group("status").
			Find(&counts)
	}
	err = m.DB.Error

	for _, r := range counts {
		switch r.Status {
		case "QUEUED":
			queued = r.Count
		case "RUNNING", "PRE_PROCESSING", "POST_PROCESSING":
			running = r.Count
		case "DONE_SUCCESSFUL":
			doneSuccessful = r.Count
		case "DONE_ERROR":
			doneError = r.Count
		case "DONE_CANCELED":
			doneCanceled = r.Count
		}
	}

	return
}

func (m *Task) List(page int, perPage int, status string) (*[]model.Task, int64, error) {
	var total int64
	var tasks = &[]model.Task{}
	if status != "" {
		m.DB.Model(&model.Task{}).Where("status = ?", status).Count(&total)
		m.DB.Order("created_at DESC").Where("status = ?", status).Limit(perPage).Offset(page * perPage).Find(&tasks)
	} else {
		m.DB.Model(&model.Task{}).Count(&total)
		m.DB.Order("created_at DESC").Limit(perPage).Offset(page * perPage).Find(&tasks)
	}

	return tasks, total, m.DB.Error
}

func (m *Task) Create(newTask *dto.NewTask, batch string, source string, session string) (*model.Task, error) {
	task := &model.Task{
		Uuid:       uuid.NewString(),
		Command:    &dto.RawResolved{Raw: newTask.Command},
		InputFile:  &dto.RawResolved{Raw: newTask.InputFile},
		OutputFile: &dto.RawResolved{Raw: newTask.OutputFile},
		Metadata:   newTask.Metadata, // Ensure Metadata is not nil
		Name:       newTask.Name,
		Priority:   newTask.Priority,
		Progress:   0,
		Source:     source,
		Status:     dto.QUEUED,
		Batch:      batch,
		Session:    session,
	}
	if newTask.PreProcessing != nil {
		task.PreProcessing = &dto.PrePostProcessing{
			ScriptPath:  &dto.RawResolved{Raw: newTask.PreProcessing.ScriptPath},
			SidecarPath: &dto.RawResolved{Raw: newTask.PreProcessing.SidecarPath},
		}
	}
	if newTask.PostProcessing != nil {
		task.PostProcessing = &dto.PrePostProcessing{
			ScriptPath:  &dto.RawResolved{Raw: newTask.PostProcessing.ScriptPath},
			SidecarPath: &dto.RawResolved{Raw: newTask.PostProcessing.SidecarPath},
		}
	}
	db := m.DB.Create(task)
	return task, db.Error
}

func (m *Task) Delete(w *model.Task) error {
	m.DB.Delete(w)
	return m.DB.Error
}

func (m *Task) First(uuid string) (*model.Task, error) {
	var task *model.Task
	db := m.DB.Where("uuid", uuid).First(&task)
	return task, db.Error
}

func (m *Task) ByBatchId(uuid string, page int, perPage int) (*[]model.Task, int64, error) {
	total, _ := m.Count()

	var tasks = &[]model.Task{}
	m.DB.Order("created_at DESC").Where("batch = ?", uuid).Limit(perPage).Offset(page * perPage).Find(&tasks)
	return tasks, total, m.DB.Error
}

func (m *Task) CountNonFinishedTasksByBatchId(uuid string) (int64, error) {
	var count int64
	db := m.DB.Model(&model.Task{}).Where("batch = ? and status != 'DONE_SUCCESSFUL' and status != 'DONE_ERROR' and status != 'DONE_CANCELED'").Count(&count)
	return count, db.Error
}

func (m *Task) Count() (int64, error) {
	var count int64
	db := m.DB.Model(&model.Task{}).Count(&count)
	return count, db.Error
}

func (m *Task) CountDeleted() (int64, error) {
	var count int64
	db := m.DB.Unscoped().Model(&model.Task{}).Where("deleted_at IS NOT NULL").Count(&count)
	return count, db.Error
}

func (m *Task) CountBySource(source string) (int64, error) {
	var count int64
	db := m.DB.Unscoped().Model(&model.Task{}).Where("deleted_at IS NOT NULL and source = ?", source).Count(&count)
	return count, db.Error
}

func (m *Task) CountDeletedByStatus(status dto.TaskStatus) (int64, error) {
	var count int64
	db := m.DB.Unscoped().Model(&model.Task{}).Where("deleted_at IS NOT NULL and status = ?", status).Count(&count)
	return count, db.Error
}

func (m *Task) CountByStatus(status dto.TaskStatus) (int64, error) {
	var count int64
	db := m.DB.Model(&model.Task{}).Where("status = ?", status).Count(&count)
	return count, db.Error
}

func (m *Task) NextQueued() (*model.Task, error) {
	var task *model.Task
	db := m.DB.Order("priority DESC, created_at ASC").Where("status", dto.QUEUED).First(&task)
	if db.RowsAffected == 0 {
		return nil, nil
	}
	return task, db.Error
}

func (m *Task) UpdateTask(task *model.Task) (*model.Task, error) {
	db := m.DB.Save(task)
	return task, db.Error
}
