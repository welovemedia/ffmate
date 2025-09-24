package repository

import (
	"errors"

	"github.com/welovemedia/ffmate/internal/database/model"
	"github.com/welovemedia/ffmate/internal/dto"
	"gorm.io/gorm"
	"goyave.dev/goyave/v5/database"
)

type Task struct {
	DB *gorm.DB
}

func (r *Task) Setup() *Task {
	r.DB.AutoMigrate(&model.Task{})
	return r
}

func (m *Task) First(uuid string) (*model.Task, error) {
	var task model.Task
	result := m.DB.Preload("Client").Preload("Labels").Where("uuid = ?", uuid).First(&task)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &task, nil
}

func (m *Task) Delete(w *model.Task) error {
	m.DB.Delete(w)
	return m.DB.Error
}

func (r *Task) List(page int, perPage int) (*[]model.Task, int64, error) {
	var tasks = &[]model.Task{}
	tx := r.DB.Preload("Client").Preload("Labels").Order("created_at DESC")
	d := database.NewPaginator(tx, page+1, perPage, tasks)
	err := d.Find()
	return d.Records, d.Total, err
}

func (r *Task) ListByBatch(uuid string, page int, perPage int) (*[]model.Task, int64, error) {
	var tasks = &[]model.Task{}
	tx := r.DB.Preload("Client").Preload("Labels").Order("created_at DESC").Where("batch = ?", uuid)
	d := database.NewPaginator(tx, page+1, perPage, tasks)
	err := d.Find()
	return d.Records, d.Total, err
}

func (r *Task) Add(newTask *model.Task) (*model.Task, error) {
	db := r.DB.Preload("Labels").Create(newTask)

	for i := range newTask.Labels {
		r.DB.FirstOrCreate(&newTask.Labels[i], model.Label{Value: newTask.Labels[i].Value})
	}

	r.DB.Model(newTask).Association("Labels").Replace(newTask.Labels)

	if db.Error != nil {
		return newTask, db.Error
	}
	return r.First(newTask.Uuid)
}

func (r *Task) Update(task *model.Task) (*model.Task, error) {
	task.Client = nil // will be re-linked during save
	db := r.DB.Session(&gorm.Session{FullSaveAssociations: true}).Preload("Labels").Save(task)
	if db.Error != nil {
		return task, db.Error
	}
	return r.First(task.Uuid)
}

func (r *Task) Count() (int64, error) {
	var count int64
	db := r.DB.Model(&model.Task{}).Count(&count)
	return count, db.Error
}

func (r *Task) CountUnfinishedByBatch(uuid string) (int64, error) {
	var count int64
	db := r.DB.Model(&model.Task{}).Where("batch = ? and status != 'DONE_SUCCESSFUL' and status != 'DONE_ERROR' and status != 'DONE_CANCELED'", uuid).Count(&count)
	return count, db.Error
}

/**
 * Stats (systray) related methods
 */

type statusCount struct {
	Status string
	Count  int
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

/**
 * Processing related methods
 */
func (m *Task) NextQueued(amount int, clientLabels dto.Labels) (*[]model.Task, error) {
	var tasks []model.Task

	err := m.DB.Transaction(func(tx *gorm.DB) error {
		// Subquery: task IDs with at least one overlapping label
		sub := tx.Model(&model.Task{}).
			Select("DISTINCT tasks.id").
			Joins("JOIN task_labels tl ON tl.task_id = tasks.id").
			Joins("JOIN labels l ON l.id = tl.label_id").
			Where("tasks.status = ?", dto.QUEUED).
			Where("l.value IN ?", clientLabels).
			Order("tasks.priority DESC, tasks.created_at ASC").
			Limit(amount)

		// Load tasks
		if err := tx.Preload("Labels").Where("tasks.id IN (?)", sub).Find(&tasks).Error; err != nil {
			return err
		}

		if len(tasks) == 0 {
			return gorm.ErrRecordNotFound
		}

		// Update selected tasks to RUNNING
		ids := make([]uint, len(tasks))
		for i, t := range tasks {
			ids[i] = t.ID
		}

		if err := tx.Model(&model.Task{}).
			Where("id IN ?", ids).
			Update("status", dto.RUNNING).Error; err != nil {
			return err
		}

		return nil
	})

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &tasks, err
}

/**
 * Stats (telemetry) related methods
 */

func (r *Task) CountAllBySource(source string) (int64, error) {
	var count int64
	db := r.DB.Model(&model.Task{}).Where("source = ?", source).Count(&count)
	return count, db.Error
}

func (r *Task) CountByStatus(status dto.TaskStatus) (int64, error) {
	var count int64
	db := r.DB.Model(&model.Task{}).Where("status = ?", status).Count(&count)
	return count, db.Error
}
func (r *Task) CountDeletedByStatus(status dto.TaskStatus) (int64, error) {
	var count int64
	db := r.DB.Unscoped().Model(&model.Task{}).Unscoped().Where("status = ? AND deleted_at IS NOT NULL", status).Count(&count)
	return count, db.Error
}

func (r *Task) CountDeleted() (int64, error) {
	var count int64
	db := r.DB.Unscoped().Model(&model.Task{}).Unscoped().Where("deleted_at IS NOT NULL").Count(&count)
	return count, db.Error
}
