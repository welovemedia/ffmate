package model

type Label struct {
	ID    uint   `gorm:"primaryKey"`
	Value string `gorm:"uniqueIndex"`
}
