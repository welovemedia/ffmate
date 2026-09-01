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
	result := r.DB.Preload("Client.Labels").Preload("Labels").Where("uuid = ?", uuid).First(&task)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &task, nil
}

func (r *Task) Delete(w *model.Task) error {
	return r.DB.Delete(w).Error
}

func (r *Task) List(page int, perPage int, status dto.TaskStatus) (*[]model.Task, int64, error) {
	var tasks = &[]model.Task{}
	tx := r.DB.Preload("Client.Labels").Preload("Labels").Order("created_at DESC")
	if status != dto.All {
		tx = tx.Where("status = ?", status)
	}
	d := database.NewPaginator(tx, page+1, perPage, tasks)
	err := d.Find()
	return d.Records, d.Total, err
}

func (r *Task) ListByBatch(uuid string, page int, perPage int) (*[]model.Task, int64, error) {
	var tasks = &[]model.Task{}
	tx := r.DB.Preload("Client.Labels").Preload("Labels").Order("created_at DESC").Where("batch = ?", uuid)
	d := database.NewPaginator(tx, page+1, perPage, tasks)
	err := d.Find()
	return d.Records, d.Total, err
}

func (r *Task) Add(newTask *model.Task) (*model.Task, error) {
	err := r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Labels").Create(newTask).Error; err != nil {
			return err
		}

		if err := upsertLabels(tx, newTask.Labels); err != nil {
			return err
		}

		return tx.Model(newTask).Association("Labels").Replace(newTask.Labels)
	})
	if err != nil {
		return newTask, err
	}
	return r.First(newTask.UUID)
}

func (r *Task) Update(task *model.Task) (*model.Task, error) {
	task.Client = nil // will be re-linked during save
	err := r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Labels").Session(&gorm.Session{FullSaveAssociations: true}).Save(task).Error; err != nil {
			return err
		}

		if err := upsertLabels(tx, task.Labels); err != nil {
			return err
		}

		return tx.Model(task).Association("Labels").Replace(task.Labels)
	})
	if err != nil {
		return task, err
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
	db := r.DB.Model(&model.Task{}).
		Where("batch = ? AND status NOT IN (?)", uuid, []dto.TaskStatus{dto.DoneSuccessful, dto.DoneError, dto.DoneCanceled}).
		Count(&count)
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

	err = r.DB.Model(&model.Task{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&counts).Error
	if err != nil {
		return
	}

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
			Select("tasks.id").
			Where("tasks.status = ?", dto.Queued).
			Where(whereSQL, whereArgs...).
			Order("tasks.priority DESC, tasks.created_at ASC").
			Limit(amount)

		if err := tx.Preload("Labels").
			Where("tasks.id IN (?)", sub).
			Order("tasks.priority DESC, tasks.created_at ASC").
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
// Uses EXISTS/NOT EXISTS against task_labels/labels instead of a LEFT JOIN so
// the outer query never produces duplicate rows per task (a JOIN + DISTINCT
// combined with ORDER BY on non-selected columns is rejected by postgres).
//
// Rules:
// - If clientLabels is empty → only unlabeled tasks are eligible
// - If task has no labels → always eligible
// - If both have labels → must match at least one pattern (clientLabel LIKE REPLACE(l.value, '*', '%'))
func buildLabelFilterSQL(clientLabels dto.Labels) (string, []any) {
	noLabels := "NOT EXISTS (SELECT 1 FROM task_labels tl WHERE tl.task_id = tasks.id)"

	if len(clientLabels) == 0 {
		return noLabels, nil
	}

	// Build "clientLabel LIKE REPLACE(l.value, '*', '%')" for each
	labelConds := make([]string, len(clientLabels))
	args := make([]any, len(clientLabels))

	for i, lbl := range clientLabels {
		labelConds[i] = "? LIKE REPLACE(l.value, '*', '%')"
		args[i] = lbl
	}

	matchesLabel := fmt.Sprintf(`EXISTS (
		SELECT 1 FROM task_labels tl
		JOIN labels l ON l.id = tl.label_id
		WHERE tl.task_id = tasks.id AND (%s)
	)`, strings.Join(labelConds, " OR "))

	sql := fmt.Sprintf("(%s OR %s)", noLabels, matchesLabel)

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
	db := r.DB.Unscoped().Model(&model.Task{}).Where("status = ? AND deleted_at IS NOT NULL", status).Count(&count)
	return count, db.Error
}

func (r *Task) CountDeleted() (int64, error) {
	var count int64
	db := r.DB.Unscoped().Model(&model.Task{}).Where("deleted_at IS NOT NULL").Count(&count)
	return count, db.Error
}
