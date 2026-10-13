package account

import (
	"context"
	"fmt"

	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

func (c *Client) DeleteInstallServiceAccounts(ctx context.Context, installID string) error {
	if installID == "" {
		return fmt.Errorf("service account cleanup requires an install")
	}
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var install struct{ ID, OrgID string }
		if err := tx.Unscoped().Model(&app.Install{}).Scopes(scopes.WithDisableViews).
			Select("id", "org_id").Where(app.Install{ID: installID}).Find(&install).Error; err != nil {
			return fmt.Errorf("load install for service account cleanup: %w", err)
		}
		if install.ID != "" {
			var org struct{ ID string }
			if err := tx.Unscoped().Model(&app.Org{}).Select("id").
				Clauses(clause.Locking{Strength: "SHARE"}).Where(app.Org{ID: install.OrgID}).Find(&org).Error; err != nil {
				return fmt.Errorf("lock org for install service account cleanup: %w", err)
			}
			if err := tx.Unscoped().Model(&app.Install{}).Scopes(scopes.WithDisableViews).
				Select("id", "org_id").Clauses(clause.Locking{Strength: "UPDATE"}).
				Where(app.Install{ID: installID, OrgID: install.OrgID}).Find(&install).Error; err != nil {
				return fmt.Errorf("lock install for service account cleanup: %w", err)
			}
		}

		var stackIDs []string
		if err := tx.Unscoped().Model(&app.InstallStack{}).Where(app.InstallStack{InstallID: installID}).
			Order("id").Pluck("id", &stackIDs).Error; err != nil {
			return fmt.Errorf("list stacks for service account cleanup: %w", err)
		}
		query := tx.Where(app.ManagedServiceAccount{OwnerType: plugins.TableName(tx, app.Install{}), OwnerID: installID})
		if len(stackIDs) != 0 {
			query = query.Or(tx.Where(app.ManagedServiceAccount{OwnerType: plugins.TableName(tx, app.InstallStack{})}).
				Where("owner_id = ANY(?)", pq.Array(stackIDs)))
		}
		if err := deleteManagedServiceAccounts(tx, query); err != nil {
			return err
		}
		return New(Params{DB: tx}).DeleteInstallStackServiceAccounts(ctx, installID)
	})
}

func (c *Client) DeleteOrgServiceAccounts(ctx context.Context, orgID string) error {
	if orgID == "" {
		return fmt.Errorf("service account cleanup requires an org")
	}
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var org struct{ ID string }
		if err := tx.Unscoped().Model(&app.Org{}).Select("id").
			Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.Org{ID: orgID}).Find(&org).Error; err != nil {
			return fmt.Errorf("lock org for service account cleanup: %w", err)
		}
		if err := deleteManagedServiceAccounts(tx, tx.Where(app.ManagedServiceAccount{OrgID: orgID})); err != nil {
			return err
		}
		var stackIDs []string
		if err := tx.Unscoped().Model(&app.InstallStack{}).Where(app.InstallStack{OrgID: orgID}).
			Order("id").Pluck("id", &stackIDs).Error; err != nil {
			return fmt.Errorf("list org stacks for service account cleanup: %w", err)
		}
		client := New(Params{DB: tx})
		for _, stackID := range stackIDs {
			if err := client.DeleteServiceAccount(ctx, stackID); err != nil {
				return err
			}
		}
		return nil
	})
}

func deleteManagedServiceAccounts(tx, query *gorm.DB) error {
	var bindings []app.ManagedServiceAccount
	if err := query.Unscoped().Order("account_id").Find(&bindings).Error; err != nil {
		return fmt.Errorf("list managed service accounts for cleanup: %w", err)
	}
	for _, binding := range bindings {
		if err := deleteServiceAccount(tx, binding.AccountID); err != nil {
			return err
		}
	}
	return nil
}
