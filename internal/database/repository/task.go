package repository

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"github.com/welovemedia/ffmate/v2/internal/dto"
	"gorm.io/gorm"
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
	result := r.DB.Preload("Client").Preload("Labels").Where("uuid = ?", uuid).First(&task)
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

func (r *Task) List(page int, perPage int, status dto.TaskStatus) (*[]model.Task, int64, error) {
	var tasks = &[]model.Task{}
	tx := r.DB.Preload("Client").Preload("Labels").Order("created_at DESC")
	if status != dto.All {
		tx = tx.Where("status = ?", status)
	}
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
		_ = r.DB.FirstOrCreate(&newTask.Labels[i], model.Label{Value: newTask.Labels[i].Value})
	}

	_ = r.DB.Model(newTask).Association("Labels").Replace(newTask.Labels)

	if db.Error != nil {
		return newTask, db.Error
	}
	return r.First(newTask.UUID)
}

func (r *Task) Update(task *model.Task) (*model.Task, error) {
	task.Client = nil // will be re-linked during save
	db := r.DB.Session(&gorm.Session{FullSaveAssociations: true}).Preload("Labels").Save(task)
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

func (r *Task) FailRunningTasksForStartingClient(identifier string) ([]model.Task, error) {
	var tasks []model.Task
	err := r.DB.Raw(`
		UPDATE tasks
		SET status = ?, error = ?, finished_at = ?,remaining = -1, progress = 100
		WHERE status = ? AND client_identifier = ?
		RETURNING *;
	`, dto.DoneError, "client disconnected during execution", time.Now().UnixMilli(), dto.Running, identifier).Scan(&tasks).Error

	return tasks, err
}

func (r *Task) FailRunningTasksForOfflineClients() ([]model.Task, error) {
	threshold := time.Now().Add(-60 * time.Second).UnixMilli() // int64

	now := time.Now().UnixMilli()
	var tasks []model.Task

	err := r.DB.Raw(`
		UPDATE tasks
		SET status = ?, error = ?, finished_at = ?, remaining = -1, progress = 100
		WHERE status = ?
		  AND client_identifier IN (
		      SELECT identifier
		      FROM client
		      WHERE last_seen < ?
		  )
		RETURNING *;
	`, dto.DoneError, "client disconnected during execution", now, dto.Running, threshold).
		Scan(&tasks).Error

	return tasks, err
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

func (r *Task) NextQueued(amount int, clientLabels dto.Labels) (*[]model.Task, error) {
	var tasks []model.Task

	err := r.DB.Transaction(func(tx *gorm.DB) error {
		whereSQL, whereArgs := buildLabelFilterSQL(clientLabels)

		sub := tx.Model(&model.Task{}).
			Select("DISTINCT tasks.id").
			Joins("LEFT JOIN task_labels tl ON tl.task_id = tasks.id").
			Joins("LEFT JOIN labels l ON l.id = tl.label_id").
			Where("tasks.status = ?", dto.Queued).
			Where(whereSQL, whereArgs...).
			Order("tasks.priority DESC, tasks.created_at ASC").
			Limit(amount)

		if err := tx.Preload("Labels").
			Where("tasks.id IN (?)", sub).
			Find(&tasks).Error; err != nil {
			return err
		}

		if len(tasks) == 0 {
			return gorm.ErrRecordNotFound
		}

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

// buildLabelFilterSQL generates the SQL WHERE clause and args for filtering
// queued tasks by client labels, respecting wildcard labels stored in the DB.
//
// Rules:
// - If clientLabels is empty → only unlabeled tasks are eligible (l.id IS NULL)
// - If task has no labels → always eligible
// - If both have labels → must match at least one pattern (clientLabel LIKE REPLACE(l.value, '*', '%'))
func buildLabelFilterSQL(clientLabels dto.Labels) (string, []any) {
	if len(clientLabels) == 0 {
		return "l.id IS NULL", nil
	}

	// Build "clientLabel LIKE REPLACE(l.value, '*', '%')" for each
	labelConds := make([]string, len(clientLabels))
	args := make([]any, len(clientLabels))

	for i, lbl := range clientLabels {
		labelConds[i] = "? LIKE REPLACE(l.value, '*', '%')"
		args[i] = lbl
	}

	// Allow tasks with no labels (l.id IS NULL) or any matching label
	sql := fmt.Sprintf("(l.id IS NULL OR (%s))", strings.Join(labelConds, " OR "))

	return sql, args
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
