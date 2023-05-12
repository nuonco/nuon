// deployment.go
package models

import (
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/services/api/internal/jobs"
	"gorm.io/gorm"
)

type Deployment struct {
	ModelV2

	ComponentID string
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

	if err := mgr.CreateDeployment(ctx, d.ID); err != nil {
		return fmt.Errorf("unable to create deployment: %w", err)
	}

	return nil
}

func (Deployment) IsNode() {}

func (d Deployment) GetID() string {
	return d.ModelV2.ID
}

func (d Deployment) GetCreatedAt() time.Time {
	return d.ModelV2.CreatedAt
}

func (d Deployment) GetUpdatedAt() time.Time {
	return d.ModelV2.UpdatedAt
}
