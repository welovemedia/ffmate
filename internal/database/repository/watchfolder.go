package repository

import (
	"errors"
	"time"

	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"goyave.dev/goyave/v5/database"
)

type Watchfolder struct {
	DB *gorm.DB
}

func (r *Watchfolder) Setup() *Watchfolder {
	_ = r.DB.AutoMigrate(&model.Watchfolder{})
	return r
}

func (r *Watchfolder) First(uuid string) (*model.Watchfolder, error) {
	var watchfolder model.Watchfolder
	result := r.DB.Preload("Labels").Where("uuid = ?", uuid).First(&watchfolder)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &watchfolder, nil
}

func (r *Watchfolder) Delete(w *model.Watchfolder) error {
	return r.DB.Delete(w).Error
}

func (r *Watchfolder) List(page int, perPage int) (*[]model.Watchfolder, int64, error) {
	var watchfolders = &[]model.Watchfolder{}

	// return all (internal usage)
	if page == -1 && perPage == -1 {
		err := r.DB.Preload("Labels").Order("created_at DESC").Find(watchfolders).Error
		return watchfolders, int64(len(*watchfolders)), err
	}

	tx := r.DB.Preload("Labels").Order("created_at DESC")
	d := database.NewPaginator(tx, page+1, perPage, watchfolders)
	err := d.Find()
	return d.Records, d.Total, err
}

func (r *Watchfolder) Save(watchfolder *model.Watchfolder) (*model.Watchfolder, error) {
	err := r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(watchfolder).Error; err != nil {
			return err
		}

		if err := upsertLabels(tx, watchfolder.Labels); err != nil {
			return err
		}

		return tx.Model(watchfolder).Association("Labels").Replace(watchfolder.Labels)
	})

	return watchfolder, err
}

func (r *Watchfolder) Count() (int64, error) {
	var count int64
	db := r.DB.Model(&model.Watchfolder{}).Count(&count)
	return count, db.Error
}

/**
 * Processing related methods
 */

func (r *Watchfolder) FirstAndLock(uuid string) (*model.Watchfolder, bool, error) {
	var watchfolder model.Watchfolder
	var locked = false
	err := r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Labels").
			Where("uuid = ?", uuid).
			First(&watchfolder).Error; err != nil {
			return err
		}

		// calculate the next allowed run time based on Interval + 50ms buffer
		nextRun := time.UnixMilli(watchfolder.LastRun).Add(time.Duration(watchfolder.Interval)*time.Second - 50*time.Millisecond)
		if time.Now().Before(nextRun) {
			locked = true
			return nil
		}

		if err := tx.Model(&watchfolder).Update("last_run", time.Now().UnixMilli()).Error; err != nil {
			return err
		}

		return nil
	})

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, locked, nil
	}

	return &watchfolder, locked, err
}

/**
 * Stats (telemetry) related methods
 */

func (r *Watchfolder) CountDeleted() (int64, error) {
	var count int64
	result := r.DB.Unscoped().Model(&model.Watchfolder{}).Where("deleted_at IS NOT NULL").Count(&count)
	return count, result.Error
}
