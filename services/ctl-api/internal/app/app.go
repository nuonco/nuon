package app

import (
	"context"

	"gorm.io/gorm"
)

const AppHooksKey string = "hooks_apps"

type AppHooks interface {
	AfterCreate(context.Context, string)
}

type App struct {
	Model

	CreatedByID string
	Name        string
	OrgID       string
	Org         Org         `faker:"-"`
	Components  []Component `faker:"-"`
	Installs    []Install   `faker:"-"`
}

func (a *App) AfterCreate(tx *gorm.DB) error {
	ctx := tx.Statement.Context
	val := ctx.Value(AppHooksKey)
	hooks, ok := val.(AppHooks)
	if !ok {
		return nil
	}

	hooks.AfterCreate(ctx, a.ID)
	return nil
}
