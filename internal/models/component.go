// component.go
package models

import (
	"time"

	"github.com/google/uuid"
)

type Component struct {
	Model

	Name  string
	AppID uuid.UUID
	App   App `fake:"skip"`

	BuildImage string `json:"container_image_url"`
	Type       string `json:"type"`

	Deployments []Deployment `fake:"skip"`
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
