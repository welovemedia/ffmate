package repository

import (
	"errors"

	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"goyave.dev/goyave/v5/database"
)

type Task struct {
	DB *gorm.DB
}

func (r *Task) Setup() *Task {
	_ = r.DB.AutoMigrate(&model.Task{})
	return r
}

func (r *Task) First(uuid string) (*model.Task, error) {
	var task model.Task
	result := r.DB.Preload("Client").Where("uuid = ?", uuid).First(&task)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &task, nil
}

func (r *Task) Delete(w *model.Task) error {
	r.DB.Delete(w)
	return r.DB.Error
}

func (r *Task) List(page int, perPage int) (*[]model.Task, int64, error) {
	var tasks = &[]model.Task{}
	tx := r.DB.Preload("Client").Order("created_at DESC")
	d := database.NewPaginator(tx, page+1, perPage, tasks)
	err := d.Find()
	return d.Records, d.Total, err
}

func (r *Task) ListByBatch(uuid string, page int, perPage int) (*[]model.Task, int64, error) {
	var tasks = &[]model.Task{}
	tx := r.DB.Preload("Client").Order("created_at DESC").Where("batch = ?", uuid)
	d := database.NewPaginator(tx, page+1, perPage, tasks)
	err := d.Find()
	return d.Records, d.Total, err
}

func (r *Task) Add(newTask *model.Task) (*model.Task, error) {
	db := r.DB.Create(newTask)
	if db.Error != nil {
		return newTask, db.Error
	}
	return r.First(newTask.UUID)
}

func (r *Task) Update(task *model.Task) (*model.Task, error) {
	task.Client = nil // will be re-linked during save
	db := r.DB.Session(&gorm.Session{FullSaveAssociations: true}).Save(task)
	if db.Error != nil {
		return task, db.Error
	}
	return r.First(task.UUID)
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

func (r *Task) CountAllStatus() (queued, running, doneSuccessful, doneError, doneCanceled int, err error) {
	var counts []statusCount

	r.DB.Model(&model.Task{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&counts)

	err = r.DB.Error

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

func (r *Task) NextQueued(amount int) (*[]model.Task, error) {
	var tasks []model.Task

	err := r.DB.Transaction(func(tx *gorm.DB) error {
		// Select tasks with FOR UPDATE
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Client").
			Order("priority DESC, created_at ASC").
			Where("status = ?", dto.Queued).
			Limit(amount).
			Find(&tasks).Error; err != nil {
			return err
		}

		if len(tasks) == 0 {
			return gorm.ErrRecordNotFound
		}

		// Extract IDs to ensure we only update the selected ones
		ids := make([]uint, len(tasks))
		for i, t := range tasks {
			ids[i] = t.ID
		}

		if err := tx.Model(&model.Task{}).
			Where("id IN ?", ids).
			Update("status", dto.Running).Error; err != nil {
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
