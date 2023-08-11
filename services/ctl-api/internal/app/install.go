package app

import (
	"context"

	"gorm.io/gorm"
)

const InstallHooksKey string = "hooks_installs"

type InstallHooks interface {
	AfterCreate(context.Context, string)
}

type Install struct {
	Model
	CreatedByID string

	Name  string
	AppID string
	App   App
}

func (i *Install) AfterCreate(tx *gorm.DB) error {
	ctx := tx.Statement.Context
	val := ctx.Value(InstallHooksKey)
	hooks, ok := val.(InstallHooks)
	if !ok {
		return nil
	}

	hooks.AfterCreate(ctx, i.ID)
	return nil
}
