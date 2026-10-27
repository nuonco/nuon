package helpers

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
)

// ProvisionJobServiceAccount creates the job SA + runner_job role (and install
// role when the runner is install-scoped). Call once when the job is created;
// claim only issues a token for this account.
func (h *Helpers) ProvisionJobServiceAccount(ctx context.Context, orgID string, job *app.RunnerJob) error {
	if orgID == "" || job == nil || job.ID == "" {
		return fmt.Errorf("job service account requires org and job")
	}

	acct, err := account.New(account.Params{DB: h.db}).CreateServiceAccount(ctx, job.ID, "runner-job")
	if err != nil {
		return fmt.Errorf("create job service account: %w", err)
	}

	binding := app.ManagedServiceAccount{
		AccountID:   acct.ID,
		OrgID:       orgID,
		OwnerType:   plugins.TableName(h.db, app.RunnerJob{}),
		OwnerID:     job.ID,
		Purpose:     app.ManagedServiceAccountPurposeRunnerJob,
		InstanceKey: "default",
	}
	if err := h.db.WithContext(ctx).Create(&binding).Error; err != nil {
		return fmt.Errorf("bind job service account: %w", err)
	}

	roleCtx := cctx.SetAccountIDContext(ctx, acct.ID)
	roleCtx = cctx.SetOrgIDContext(roleCtx, orgID)
	role := &app.Role{
		OrgID:       generics.NewNullString(orgID),
		RoleType:    app.RoleTypeRunnerJob,
		Version:     app.RoleVersionV2,
		Title:       "Runner job",
		Description: "Scoped access to one runner job.",
		Policies: []app.Policy{{
			OrgID: generics.NewNullString(orgID),
			Name:  app.PolicyNameRunnerJob,
			Permissions: pgtype.Hstore{
				permissions.APIPermissionKey:                                 generics.ToPtr(string(permissions.APIRunner)),
				permissions.Object(orgID, permissions.KindRunnerJob, job.ID): permissions.PermissionAll.ToStrPtr(),
			},
		}},
	}
	if err := authz.New(authz.Params{DB: h.db}).EnsureManagedServiceAccountRole(roleCtx, acct.ID, role); err != nil {
		return fmt.Errorf("create job role: %w", err)
	}

	if job.RunnerID == "" {
		return nil
	}
	var runner app.Runner
	if err := h.db.WithContext(ctx).Preload("RunnerGroup").Where(app.Runner{ID: job.RunnerID}).Take(&runner).Error; err != nil {
		return nil
	}
	if runner.RunnerGroup.Type == app.RunnerGroupTypeInstall && runner.RunnerGroup.OwnerID != "" {
		if err := authz.New(authz.Params{DB: h.db}).EnsureRunnerInstallRole(ctx, orgID, runner.RunnerGroup.OwnerID, acct.ID); err != nil {
			return fmt.Errorf("ensure job install role: %w", err)
		}
	}
	return nil
}
