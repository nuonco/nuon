package app

import "gorm.io/datatypes"

type Component struct {
	Model
	Name        string
	AppID       string
	App         App `faker:"-"`
	CreatedByID string
	Config      datatypes.JSON `gorm:"not null;default:'{}'" json:"config"`
}
