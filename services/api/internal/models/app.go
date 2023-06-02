package models

import (
	"fmt"
	"time"

	appsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/apps/v1"
	"github.com/powertoolsdev/mono/services/api/internal/jobs"
	"gorm.io/gorm"
)

type App struct {
	Model
	CreatedByID string
	Name        string
	OrgID       string
	Org         Org         `faker:"-"`
	Components  []Component `faker:"-"`
	Installs    []Install   `faker:"-"`
}

func (a *App) AfterCreate(tx *gorm.DB) (err error) {
	ctx := tx.Statement.Context
	mgr, err := jobs.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("unable to get job manager: %w", err)
	}

	if err := mgr.CreateApp(ctx, a.ID); err != nil {
		return fmt.Errorf("unable to create app: %w", err)
	}

	return nil
}

func (App) IsNode() {}

func (a App) GetID() string {
	return a.Model.ID
}

func (a App) GetCreatedAt() time.Time {
	return a.Model.CreatedAt
}

func (a App) GetUpdatedAt() time.Time {
	return a.Model.UpdatedAt
}

func (a App) ToProvisionRequest() *appsv1.ProvisionRequest {
	return &appsv1.ProvisionRequest{
		OrgId: a.OrgID,
		AppId: a.ID,
	}
}
