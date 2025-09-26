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
	r.DB.AutoMigrate(&model.Preset{})
	return r
}

func (m *Preset) First(uuid string) (*model.Preset, error) {
	var preset model.Preset
	result := m.DB.Preload("Labels").Where("uuid = ?", uuid).First(&preset)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &preset, nil
}

func (m *Preset) Delete(w *model.Preset) error {
	m.DB.Delete(w)
	return m.DB.Error
}

func (r *Preset) List(page int, perPage int) (*[]model.Preset, int64, error) {
	var presets = &[]model.Preset{}
	tx := r.DB.Preload("Labels").Order("created_at DESC")
	d := database.NewPaginator(tx, page+1, perPage, presets)
	err := d.Find()
	return d.Records, d.Total, err
}

func (r *Preset) Save(preset *model.Preset) (*model.Preset, error) {
	db := r.DB.Save(preset)

	for i := range preset.Labels {
		r.DB.FirstOrCreate(&preset.Labels[i], model.Label{Value: preset.Labels[i].Value})
	}

	r.DB.Model(preset).Association("Labels").Replace(preset.Labels)

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
