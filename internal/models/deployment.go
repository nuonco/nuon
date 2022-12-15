// deployment.go
package models

import (
	"time"

	"github.com/google/uuid"
)

type Deployment struct {
	Model

	ComponentID uuid.UUID
	Component   Component `fake:"skip"`
}

func (Deployment) IsNode() {}

func (d Deployment) GetID() string {
	return d.Model.ID.String()
}

func (d Deployment) GetCreatedAt() time.Time {
	return d.Model.CreatedAt
}

func (d Deployment) GetUpdatedAt() time.Time {
	return d.Model.UpdatedAt
}
