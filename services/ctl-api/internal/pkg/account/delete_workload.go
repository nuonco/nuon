package account

import (
	"context"
	"fmt"

	"github.com/lib/pq"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
)

func (c *Client) DeleteRunnerJobCredentials(ctx context.Context, jobID string) error {
	if jobID == "" {
		return gorm.ErrRecordNotFound
	}
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return deleteJobCredentials(tx, tx.Unscoped().Model(&app.RunnerJob{}).Select("id").Where(app.RunnerJob{ID: jobID}))
	})
}

func (c *Client) DeleteRunnerServiceAccounts(ctx context.Context, runnerID string) error {
	if runnerID == "" {
		return gorm.ErrRecordNotFound
	}
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var runner app.Runner
		if err := tx.Unscoped().Where(app.Runner{ID: runnerID}).Find(&runner).Error; err != nil {
			return err
		}
		if runner.ID != "" {
			if err := lockManagedAccountOwner(tx.Unscoped(), ManagedServiceAccountRequest{OrgID: runner.OrgID, OwnerType: plugins.TableName(tx, app.Runner{}), OwnerID: runnerID}); err != nil {
				return err
			}
		}
		if err := deleteJobCredentials(tx, tx.Unscoped().Model(&app.RunnerJob{}).Select("id").Where(app.RunnerJob{RunnerID: runnerID})); err != nil {
			return err
		}
		return c.withDB(tx).DeleteServiceAccount(ctx, runnerID)
	})
}

// deleteJobCredentials deletes the accounts and job-sourced tokens of the
// selected jobs. It locks only jobs that still hold credentials, so its cost
// follows live credentials rather than job history.
func deleteJobCredentials(tx, jobs *gorm.DB) error {
	var accts []app.Account
	if err := tx.Select("id", "subject").
		Where("account_type = ? AND subject IN (?)", app.AccountTypeService, jobs).Find(&accts).Error; err != nil {
		return fmt.Errorf("load job accounts: %w", err)
	}
	var jobIDs []string
	if err := tx.Model(&app.Token{}).Distinct("source_id").
		Where("source_type = ? AND source_id IN (?)", app.TokenSourceTypeRunnerJob, jobs).Pluck("source_id", &jobIDs).Error; err != nil {
		return fmt.Errorf("load job-sourced tokens: %w", err)
	}
	accountIDs := make([]string, 0, len(accts))
	for _, acct := range accts {
		accountIDs = append(accountIDs, acct.ID)
		jobIDs = append(jobIDs, acct.Subject)
	}
	if len(jobIDs) == 0 {
		return nil
	}
	if err := tx.Unscoped().Model(&app.RunnerJob{}).Select("id").Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ANY(?)", pq.Array(jobIDs)).Order("id").Find(&[]app.RunnerJob{}).Error; err != nil {
		return fmt.Errorf("lock jobs for credential cleanup: %w", err)
	}
	if len(accountIDs) != 0 {
		if err := deleteServiceAccounts(tx, accountIDs); err != nil {
			return err
		}
	}
	return RevokeDerivedCredentials(tx, app.TokenSourceTypeRunnerJob, jobIDs)
}

func (c *Client) DeleteTelemetryCollector(ctx context.Context, orgID, installID string) error {
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		req := ManagedServiceAccountRequest{OrgID: orgID, OwnerType: plugins.TableName(tx, app.Install{}), OwnerID: installID}
		if err := lockManagedAccountOwner(tx, req); err != nil {
			return err
		}
		return deleteManagedServiceAccounts(tx, tx.Where(app.ManagedServiceAccount{OrgID: orgID, OwnerType: req.OwnerType, OwnerID: installID, Purpose: app.ManagedServiceAccountPurposeTelemetryCollector}))
	})
}
