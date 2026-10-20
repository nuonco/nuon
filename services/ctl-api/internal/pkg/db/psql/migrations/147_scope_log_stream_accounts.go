package migrations

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
)

func (m *Migrations) Migration147ScopeLogStreamAccounts(ctx context.Context, db *gorm.DB) error {
	const batchSize = 100

	client := authz.New(authz.Params{DB: db})
	lastID := ""
	for {
		var streams []app.LogStream
		if err := db.WithContext(ctx).Unscoped().
			Select("id", "org_id").
			Where("id > ?", lastID).
			Order("id").
			Limit(batchSize).
			Find(&streams).Error; err != nil {
			return fmt.Errorf("load log streams for account scoping: %w", err)
		}
		if len(streams) == 0 {
			return nil
		}

		for _, stream := range streams {
			var acct app.Account
			err := db.WithContext(ctx).
				Select("id").
				Where(app.Account{Subject: stream.ID, AccountType: app.AccountTypeService}).
				Take(&acct).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			if err != nil {
				return fmt.Errorf("load log stream account %s: %w", stream.ID, err)
			}
			if err := client.EnsureLogStreamWriteRole(ctx, stream.OrgID, stream.ID, acct.ID); err != nil {
				return fmt.Errorf("scope log stream account %s: %w", stream.ID, err)
			}
		}
		lastID = streams[len(streams)-1].ID
	}
}
