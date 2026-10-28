package migrations

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
)

func (m *Migrations) Migration151BackfillLogStreamRoles(ctx context.Context, db *gorm.DB) error {
	const batchSize = 100

	authzClient := authz.New(authz.Params{DB: db})
	lastID := ""
	for {
		var streams []app.LogStream
		if err := db.WithContext(ctx).
			Select("id", "org_id").
			Where("deleted_at = 0").
			Where("id > ?", lastID).
			Order("id").
			Limit(batchSize).
			Find(&streams).Error; err != nil {
			return fmt.Errorf("load log streams: %w", err)
		}
		if len(streams) == 0 {
			return nil
		}
		for i := range streams {
			if err := backfillLogStreamRole(ctx, db, authzClient, &streams[i]); err != nil {
				return err
			}
		}
		lastID = streams[len(streams)-1].ID
	}
}

func backfillLogStreamRole(ctx context.Context, db *gorm.DB, authzClient *authz.Client, stream *app.LogStream) error {
	var acct app.Account
	err := db.WithContext(ctx).Preload("Roles").
		Where(app.Account{Subject: stream.ID, AccountType: app.AccountTypeService}).
		Take(&acct).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("load log stream account %s: %w", stream.ID, err)
	}

	for _, role := range acct.Roles {
		if role.RoleType == app.RoleTypeLogStream {
			return nil
		}
	}

	if err := authzClient.EnsureLogStreamRole(ctx, stream.OrgID, stream.ID, acct.ID); err != nil {
		return fmt.Errorf("grant log stream role for %s: %w", stream.ID, err)
	}
	if err := authzClient.RemoveAccountOrgRoleByType(ctx, app.RoleTypeRunner, stream.OrgID, acct.ID); err != nil {
		return fmt.Errorf("remove runner role for %s: %w", stream.ID, err)
	}
	return nil
}
