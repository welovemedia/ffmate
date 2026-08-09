package repository

import (
	"github.com/welovemedia/ffmate/v2/internal/database/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// upsertLabels ensures every label in labels exists in the database and backfills
// their IDs, using at most two queries regardless of how many labels are passed
// (as opposed to one FirstOrCreate round-trip per label).
func upsertLabels(db *gorm.DB, labels []model.Label) error {
	if len(labels) == 0 {
		return nil
	}

	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&labels).Error; err != nil {
		return err
	}

	values := make([]string, len(labels))
	for i, l := range labels {
		values[i] = l.Value
	}

	var existing []model.Label
	if err := db.Where("value IN ?", values).Find(&existing).Error; err != nil {
		return err
	}

	idByValue := make(map[string]uint, len(existing))
	for _, l := range existing {
		idByValue[l.Value] = l.ID
	}
	for i := range labels {
		labels[i].ID = idByValue[labels[i].Value]
	}

	return nil
}
