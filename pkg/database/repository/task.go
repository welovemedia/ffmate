package repository

import (
	"github.com/google/uuid"
	"github.com/welovemedia/ffmate/pkg/database/model"
	"github.com/welovemedia/ffmate/pkg/dto"
	"gorm.io/gorm"
)

type Task struct {
	DB *gorm.DB
}

func (t *Task) Setup() {
	t.DB.AutoMigrate(&model.Task{})
}

func (m *Task) List() (*[]model.Task, error) {
	var tasks = &[]model.Task{}
	m.DB.Order("created_at DESC").Find(&tasks)
	return tasks, m.DB.Error
}

func (m *Task) Create(command string, inputField string, outputFile string) (*model.Task, error) {
	task := &model.Task{Uuid: uuid.NewString(), Command: command, InputFile: inputField, OutputFile: outputFile, Progress: 0, Status: dto.QUEUED}
	db := m.DB.Create(task)
	return task, db.Error
}

func (m *Task) First(uuid string) (*model.Task, error) {
	var task *model.Task
	db := m.DB.Where("uuid", uuid).First(&task)
	return task, db.Error
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
	db := m.DB.Order("created_at ASC").Where("status", dto.QUEUED).First(&task)
	if db.RowsAffected == 0 {
		return nil, nil
	}
	return task, db.Error
}

func (m *Task) SetTaskStatus(task *model.Task, taskStatus dto.TaskStatus) (*model.Task, error) {
	task.Status = taskStatus
	db := m.DB.Save(task)
	return task, db.Error
}
