// component.go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Component struct {
	Model
	Name        string
	AppID       uuid.UUID
	App         App `faker:"-"`
	CreatedByID string
	Config      datatypes.JSON `gorm:"not null;default:'{}'" json:"config"`
	Deployments []Deployment   `faker:"-"`
}

func (Component) IsNode() {}

func (c Component) GetID() string {
	return c.Model.ID.String()
}

func (c Component) GetCreatedAt() time.Time {
	return c.Model.CreatedAt
}

func (c Component) GetUpdatedAt() time.Time {
	return c.Model.UpdatedAt
}
