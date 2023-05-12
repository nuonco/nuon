// component.go
package models

import (
	"time"

	"gorm.io/datatypes"
)

type Component struct {
	ModelV2
	Name        string
	AppID       string
	App         App `faker:"-"`
	CreatedByID string
	Config      datatypes.JSON `gorm:"not null;default:'{}'" json:"config"`
	Deployments []Deployment   `faker:"-"`
}

func (Component) IsNode() {}

func (c Component) GetID() string {
	return c.ModelV2.ID
}

func (c Component) GetCreatedAt() time.Time {
	return c.ModelV2.CreatedAt
}

func (c Component) GetUpdatedAt() time.Time {
	return c.ModelV2.UpdatedAt
}
