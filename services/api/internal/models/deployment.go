// deployment.go
package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/powertoolsdev/mono/services/api/internal/jobs"
	"gorm.io/gorm"
)

type Deployment struct {
	Model

	ComponentID uuid.UUID
	Component   Component `fake:"skip"`
	CreatedByID string

	CommitHash   string `json:"commit_hash"`
	CommitAuthor string `json:"commit_author"`
}

func (d Deployment) AfterCreate(tx *gorm.DB) error {
	ctx := tx.Statement.Context
	mgr, err := jobs.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("unable to get job manager: %w", err)
	}

	if err := mgr.CreateDeployment(ctx, d.ID.String()); err != nil {
		return fmt.Errorf("unable to create org: %w", err)
	}

	return nil
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
