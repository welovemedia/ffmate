package repository

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/welovemedia/ffmate/internal/database/model"
	"github.com/welovemedia/ffmate/internal/dto"
	"gorm.io/gorm"
)

type Task struct {
	DB *gorm.DB
}

func (t *Task) Setup() {
	err := t.DB.AutoMigrate(&model.Task{})
	if err != nil {
		fmt.Printf("failed to initialize database: %v", err)
	}
}

func (m *Task) List() (*[]model.Task, error) {
	var tasks = &[]model.Task{}
	m.DB.Order("created_at DESC").Find(&tasks)
	return tasks, m.DB.Error
}

func (m *Task) Create(newTask *dto.NewTask, batch string) (*model.Task, error) {
	task := &model.Task{
		Uuid:           uuid.NewString(),
		Command:        newTask.Command,
		InputFile:      newTask.InputFile,
		OutputFile:     newTask.OutputFile,
		Name:           newTask.Name,
		Priority:       newTask.Priority,
		Progress:       0,
		Status:         dto.QUEUED,
		Batch:          batch,
		PostProcessing: newTask.PostProcessing}
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

func (m *Task) ByBatchId(uuid string) (*[]model.Task, error) {
	var tasks = &[]model.Task{}
	m.DB.Order("created_at DESC").Where("batch = ?", uuid).Find(&tasks)
	return tasks, m.DB.Error
}

func (m *Task) Count() (int64, error) {
	var count int64
	db := m.DB.Model(&model.Task{}).Count(&count)
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
