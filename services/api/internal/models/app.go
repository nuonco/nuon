package models

import (
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/services/api/internal/jobs"
	"gorm.io/gorm"
)

type App struct {
	ModelV2
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
	return a.ModelV2.ID
}

func (a App) GetCreatedAt() time.Time {
	return a.ModelV2.CreatedAt
}

func (a App) GetUpdatedAt() time.Time {
	return a.ModelV2.UpdatedAt
}
