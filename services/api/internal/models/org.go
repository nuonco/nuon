package models

import (
	"fmt"
	"time"

	"github.com/go-playground/validator"
	"github.com/powertoolsdev/mono/pkg/clients/temporal"
	"github.com/powertoolsdev/mono/services/api/internal/jobs"
	"gorm.io/gorm"
)

type Org struct {
	Model

	CreatedByID string
	Name        string `gorm:"uniqueIndex"`
	Apps        []App  `faker:"-"`
	IsNew       bool   `gorm:"-:all"`
}

func (o *Org) AfterCreate(tx *gorm.DB) (err error) {
	ctx := tx.Statement.Context
	val := ctx.Value(temporal.ContextKey{})
	temporalClient, ok := val.(temporal.Client)
	if !ok {
		return fmt.Errorf("no temporal client configured in context: %w", err)
	}

	v := validator.New()
	mgr, err := jobs.New(v, jobs.WithClient(temporalClient))
	if err != nil {
		return fmt.Errorf("unable to get manager: %w", err)
	}

	if err := mgr.CreateOrg(ctx, o.ID.String()); err != nil {
		return fmt.Errorf("unable to create org: %w", err)
	}

	return nil
}

func (Org) IsNode() {}

func (o Org) GetID() string {
	return o.Model.ID.String()
}

func (o Org) GetCreatedAt() time.Time {
	return o.Model.CreatedAt
}

func (o Org) GetUpdatedAt() time.Time {
	return o.Model.UpdatedAt
}
