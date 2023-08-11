package app

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

const OrgHooksKey string = "hooks_orgs"

type OrgHooks interface {
	AfterCreate(context.Context, string)
}

type Org struct {
	Model

	CreatedByID     string
	Name            string `gorm:"uniqueIndex"`
	Apps            []App  `faker:"-"`
	IsNew           bool   `gorm:"-:all"`
	GithubInstallID string
}

func (o *Org) BeforeCreate(tx *gorm.DB) error {
	o.ID = domains.NewOrgID()
	return nil
}

func (o *Org) AfterCreate(tx *gorm.DB) error {
	ctx := tx.Statement.Context
	val := ctx.Value(OrgHooksKey)
	orgHooks, ok := val.(OrgHooks)
	if !ok {
		return nil
	}

	orgHooks.AfterCreate(ctx, o.ID)
	return nil
}
