package account

import (
	"context"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

func (c *Client) EnsureRunner(ctx context.Context, runnerID string) (*app.Account, error) {
	if runnerID == "" {
		return nil, fmt.Errorf("runner ID is required")
	}
	var acct *app.Account
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var runner app.Runner
		if err := tx.Where(app.Runner{ID: runnerID}).Take(&runner).Error; err != nil {
			return err
		}
		client := c.withDB(tx)
		binding, err := client.EnsureManagedServiceAccount(ctx, ManagedServiceAccountRequest{
			OrgID: runner.OrgID, OwnerType: plugins.TableName(tx, app.Runner{}), OwnerID: runner.ID,
			Purpose: app.ManagedServiceAccountPurposeRunner, InstanceKey: "default", Subject: runner.ID,
		})
		if err != nil {
			return err
		}
		if err := client.authzClient.AddAccountOrgRole(ctx, app.RoleTypeRunner, runner.OrgID, binding.AccountID); err != nil {
			return err
		}
		installID, err := client.runnerInstallID(ctx, runner.OrgID, runner.ID)
		if err != nil {
			return err
		}
		if installID != "" {
			if err := client.authzClient.EnsureRunnerInstallRole(ctx, runner.OrgID, installID, binding.AccountID); err != nil {
				return err
			}
		}
		acct, err = client.FindAccount(ctx, binding.AccountID)
		return err
	})
	return acct, err
}

func (c *Client) RestoreRunner(ctx context.Context, runnerID string) error {
	if runnerID == "" {
		return fmt.Errorf("runner ID is required")
	}
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var runner app.Runner
		if err := tx.Where(app.Runner{ID: runnerID}).Take(&runner).Error; err != nil {
			return err
		}
		req := ManagedServiceAccountRequest{
			OrgID: runner.OrgID, OwnerType: plugins.TableName(tx, app.Runner{}), OwnerID: runner.ID,
			Purpose: app.ManagedServiceAccountPurposeRunner, InstanceKey: "default", Subject: runner.ID,
		}
		if err := lockManagedAccountOwner(tx, req); err != nil {
			return err
		}
		client := c.withDB(tx)
		if _, err := client.restoreServiceAccount(ctx, req); err != nil {
			return err
		}
		_, err := client.EnsureRunner(ctx, runnerID)
		return err
	})
}

func (c *Client) CreateRunnerJob(ctx context.Context, job *app.RunnerJob) error {
	if job == nil {
		return fmt.Errorf("job is required")
	}
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if job.RunnerID == "" {
			orgID, err := cctx.OrgIDFromContext(ctx)
			if err != nil {
				return err
			}
			if err := lockOrg(tx, orgID); err != nil {
				return err
			}
			job.OrgID = orgID
			if err := tx.Omit("RunnerID").Create(job).Error; err != nil {
				return err
			}
			_, err = c.withDB(tx).EnsureRunnerJob(ctx, job.ID)
			return err
		}
		var runner app.Runner
		if err := tx.Where(app.Runner{ID: job.RunnerID}).Take(&runner).Error; err != nil {
			return err
		}
		if job.OrgID != "" && job.OrgID != runner.OrgID {
			return fmt.Errorf("job and runner belong to different orgs")
		}
		if err := validateContextOrg(ctx, runner.OrgID); err != nil {
			return err
		}
		if err := lockOrg(tx, runner.OrgID); err != nil {
			return err
		}
		if err := lockRunner(tx, runner.OrgID, runner.ID, "KEY SHARE"); err != nil {
			return err
		}
		job.OrgID = runner.OrgID
		if err := tx.Create(job).Error; err != nil {
			return err
		}
		_, err := c.withDB(tx).EnsureRunnerJob(ctx, job.ID)
		return err
	})
}

func (c *Client) EnsureRunnerJob(ctx context.Context, jobID string) (*app.ManagedServiceAccount, error) {
	if jobID == "" {
		return nil, fmt.Errorf("job ID is required")
	}
	var binding *app.ManagedServiceAccount
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var job app.RunnerJob
		if err := tx.Where(app.RunnerJob{ID: jobID}).Take(&job).Error; err != nil {
			return err
		}
		client := c.withDB(tx)
		var err error
		binding, err = client.EnsureManagedServiceAccount(ctx, ManagedServiceAccountRequest{
			OrgID: job.OrgID, OwnerType: plugins.TableName(tx, app.RunnerJob{}), OwnerID: job.ID,
			Purpose: app.ManagedServiceAccountPurposeRunnerJob, InstanceKey: "default", Subject: job.ID, Name: "runner-job",
		})
		if err != nil {
			return err
		}
		installID, err := client.runnerInstallID(ctx, job.OrgID, job.RunnerID)
		if err != nil {
			return err
		}
		return client.authzClient.EnsureRunnerJobRole(ctx, job.OrgID, job.ID, installID, binding.AccountID)
	})
	return binding, err
}

// runnerInstallID returns the install a runner serves, or "" for runners that
// are not in an install's runner group.
func (c *Client) runnerInstallID(ctx context.Context, orgID, runnerID string) (string, error) {
	if runnerID == "" {
		return "", nil
	}
	var runner app.Runner
	if err := c.db.WithContext(ctx).Preload("RunnerGroup").Where(app.Runner{ID: runnerID, OrgID: orgID}).Take(&runner).Error; err != nil {
		return "", err
	}
	if runner.RunnerGroup.Type == app.RunnerGroupTypeInstall {
		return runner.RunnerGroup.OwnerID, nil
	}
	return "", nil
}

func (c *Client) EnsureStack(ctx context.Context, stackID string) (*app.ManagedServiceAccount, error) {
	if stackID == "" {
		return nil, fmt.Errorf("stack ID is required")
	}
	var binding *app.ManagedServiceAccount
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var stack app.InstallStack
		if err := tx.Where(app.InstallStack{ID: stackID}).Take(&stack).Error; err != nil {
			return err
		}
		client := c.withDB(tx)
		var err error
		binding, err = client.EnsureManagedServiceAccount(ctx, ManagedServiceAccountRequest{
			OrgID: stack.OrgID, OwnerType: plugins.TableName(tx, app.InstallStack{}), OwnerID: stack.ID,
			Purpose: app.ManagedServiceAccountPurposeStack, InstanceKey: "default", Subject: stack.ID,
		})
		if err != nil {
			return err
		}
		if err := client.authzClient.EnsureStackInstallRole(ctx, stack.OrgID, stack.InstallID, binding.AccountID); err != nil {
			return err
		}
		return client.authzClient.RemoveAccountOrgRoleByType(ctx, app.RoleTypeOrgAdmin, stack.OrgID, binding.AccountID)
	})
	return binding, err
}

func (c *Client) RestoreStack(ctx context.Context, stackID string) error {
	if stackID == "" {
		return fmt.Errorf("stack ID is required")
	}
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var stack app.InstallStack
		if err := tx.Where(app.InstallStack{ID: stackID}).Take(&stack).Error; err != nil {
			return err
		}
		req := ManagedServiceAccountRequest{
			OrgID: stack.OrgID, OwnerType: plugins.TableName(tx, app.InstallStack{}), OwnerID: stack.ID,
			Purpose: app.ManagedServiceAccountPurposeStack, InstanceKey: "default", Subject: stack.ID,
		}
		if err := lockManagedAccountOwner(tx, req); err != nil {
			return err
		}
		client := c.withDB(tx)
		if _, err := client.restoreServiceAccount(ctx, req); err != nil {
			return err
		}
		_, err := client.EnsureStack(ctx, stackID)
		return err
	})
}

func (c *Client) RestoreInstallServiceAccounts(ctx context.Context, installID string) error {
	if installID == "" {
		return fmt.Errorf("install ID is required")
	}
	return c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var install app.Install
		if err := tx.Scopes(scopes.WithDisableViews).Where(app.Install{ID: installID}).Take(&install).Error; err != nil {
			return err
		}
		if err := lockManagedAccountOwner(tx, ManagedServiceAccountRequest{
			OrgID: install.OrgID, OwnerType: plugins.TableName(tx, app.Install{}), OwnerID: install.ID,
		}); err != nil {
			return err
		}
		var runnerIDs []string
		if err := tx.Model(&app.Runner{}).Where(app.Runner{OrgID: install.OrgID}).
			Where("runner_group_id IN (?)", tx.Model(&app.RunnerGroup{}).Select("id").Where(app.RunnerGroup{
				OrgID: install.OrgID, OwnerType: plugins.TableName(tx, app.Install{}), OwnerID: installID, Type: app.RunnerGroupTypeInstall,
			})).Order("id").Pluck("id", &runnerIDs).Error; err != nil {
			return err
		}
		client := c.withDB(tx)
		for _, runnerID := range runnerIDs {
			if err := client.RestoreRunner(ctx, runnerID); err != nil {
				return err
			}
		}
		var stackIDs []string
		if err := tx.Model(&app.InstallStack{}).Where(app.InstallStack{OrgID: install.OrgID, InstallID: installID}).
			Order("id").Pluck("id", &stackIDs).Error; err != nil {
			return err
		}
		for _, stackID := range stackIDs {
			if err := client.RestoreStack(ctx, stackID); err != nil {
				return err
			}
		}
		return nil
	})
}

func (c *Client) EnsureTelemetryCollector(ctx context.Context, orgID, installID string) (*app.ManagedServiceAccount, error) {
	var binding *app.ManagedServiceAccount
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		client := c.withDB(tx)
		var err error
		binding, err = client.EnsureManagedServiceAccount(ctx, ManagedServiceAccountRequest{
			OrgID: orgID, OwnerType: plugins.TableName(tx, app.Install{}), OwnerID: installID,
			Purpose: app.ManagedServiceAccountPurposeTelemetryCollector, InstanceKey: "default", Name: "Telemetry collector",
		})
		if err != nil {
			return err
		}
		return client.authzClient.EnsureTelemetryCollectorRole(ctx, orgID, installID, binding.AccountID)
	})
	return binding, err
}

func lockRunnerOwner(tx *gorm.DB, req ManagedServiceAccountRequest) error {
	runnerID := req.OwnerID
	jobOwner := req.OwnerType == plugins.TableName(tx, app.RunnerJob{})
	if jobOwner {
		var job app.RunnerJob
		if err := tx.Select("id", "runner_id").Where(app.RunnerJob{ID: req.OwnerID, OrgID: req.OrgID}).Take(&job).Error; err != nil {
			return err
		}
		runnerID = job.RunnerID
	}
	if runnerID != "" {
		strength := "UPDATE"
		if jobOwner {
			strength = "KEY SHARE"
		}
		if err := lockRunner(tx, req.OrgID, runnerID, strength); err != nil {
			return err
		}
	}
	if jobOwner {
		var job app.RunnerJob
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.RunnerJob{ID: req.OwnerID, OrgID: req.OrgID}).Take(&job).Error; err != nil {
			return err
		}
		if job.RunnerID != runnerID {
			return fmt.Errorf("job runner changed during provisioning")
		}
	}
	return nil
}

// lockRunner locks a runner, after its install with KEY SHARE when it serves one.
// The caller holds the org lock.
func lockRunner(tx *gorm.DB, orgID, runnerID, strength string) error {
	var runner app.Runner
	if err := tx.Preload("RunnerGroup").Where(app.Runner{ID: runnerID, OrgID: orgID}).Take(&runner).Error; err != nil {
		return err
	}
	if runner.RunnerGroup.OrgID != orgID {
		return fmt.Errorf("runner group belongs to another org")
	}
	if runner.RunnerGroup.Type == app.RunnerGroupTypeInstall {
		if err := lockInstall(tx, orgID, runner.RunnerGroup.OwnerID, "KEY SHARE"); err != nil {
			return err
		}
	}
	return tx.Select("id").Clauses(clause.Locking{Strength: strength}).
		Where(app.Runner{ID: runnerID, OrgID: orgID, RunnerGroupID: runner.RunnerGroupID}).Take(&app.Runner{}).Error
}

func validateContextOrg(ctx context.Context, orgID string) error {
	if requested, err := cctx.OrgIDFromContext(ctx); err == nil && requested != orgID {
		return fmt.Errorf("resource belongs to another org")
	}
	return nil
}
