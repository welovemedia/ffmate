package repository

import (
	"github.com/google/uuid"
	"github.com/welovemedia/ffmate/pkg/database/model"
	"gorm.io/gorm"
)

type Preset struct {
	DB *gorm.DB
}

func (t *Preset) Setup() {
	t.DB.AutoMigrate(&model.Preset{})
}

func (m *Preset) List() (*[]model.Preset, error) {
	var presets = &[]model.Preset{}
	m.DB.Order("created_at DESC").Find(&presets)
	return presets, m.DB.Error
}

func (m *Preset) Delete(w *model.Preset) error {
	m.DB.Delete(w)
	return m.DB.Error
}

func (m *Preset) Create(command string, name string, description string) (*model.Preset, error) {
	preset := &model.Preset{Uuid: uuid.NewString(), Command: command, Name: name, Description: description}
	db := m.DB.Create(preset)
	return preset, db.Error
}

func (m *Preset) First(uuid string) (*model.Preset, error) {
	var preset *model.Preset
	db := m.DB.Where("uuid", uuid).First(&preset)
	return preset, db.Error
}

func (m *Preset) FirstByName(name string) (*model.Preset, error) {
	var preset *model.Preset
	db := m.DB.Where("name", name).First(&preset)
	return preset, db.Error
}

func (m *Preset) Count() (int64, error) {
	var count int64
	db := m.DB.Model(&model.Preset{}).Count(&count)
	return count, db.Error
}
