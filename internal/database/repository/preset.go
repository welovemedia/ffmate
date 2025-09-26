package repository

import (
	"errors"

	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"gorm.io/gorm"
	"goyave.dev/goyave/v5/database"
)

type Preset struct {
	DB *gorm.DB
}

func (r *Preset) Setup() *Preset {
	_ = r.DB.AutoMigrate(&model.Preset{})
	return r
}

func (r *Preset) First(uuid string) (*model.Preset, error) {
	var preset model.Preset
	result := r.DB.Where("uuid = ?", uuid).First(&preset)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &preset, nil
}

func (r *Preset) Delete(w *model.Preset) error {
	r.DB.Delete(w)
	return r.DB.Error
}

func (r *Preset) List(page int, perPage int) (*[]model.Preset, int64, error) {
	var presets = &[]model.Preset{}
	tx := r.DB.Order("created_at DESC")
	d := database.NewPaginator(tx, page+1, perPage, presets)
	err := d.Find()
	return d.Records, d.Total, err
}

func (r *Preset) Add(newPreset *model.Preset) (*model.Preset, error) {
	db := r.DB.Create(newPreset)
	return newPreset, db.Error
}

func (r *Preset) Update(preset *model.Preset) (*model.Preset, error) {
	db := r.DB.Save(preset)
	return preset, db.Error
}

func (r *Preset) Count() (int64, error) {
	var count int64
	db := r.DB.Model(&model.Preset{}).Count(&count)
	return count, db.Error
}

/**
 * Stats (telemetry) related methods
 */

func (r *Preset) CountDeleted() (int64, error) {
	var count int64
	result := r.DB.Unscoped().Model(&model.Preset{}).Unscoped().Where("deleted_at IS NOT NULL").Count(&count)
	return count, result.Error
}
