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

// ProvisionJobServiceAccount creates the job SA + runner_job role. For an
// install-scoped runner the role also covers that install's execution state and
// runner config. Call once when the job is created; claim only issues a token
// for this account.
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

	grants := pgtype.Hstore{
		permissions.APIPermissionKey:                                 generics.ToPtr(string(permissions.APIRunner)),
		permissions.Object(orgID, permissions.KindRunnerJob, job.ID): permissions.PermissionAll.ToStrPtr(),
	}
	if job.RunnerID != "" {
		var runner app.Runner
		if err := h.db.WithContext(ctx).Preload("RunnerGroup").Where(app.Runner{ID: job.RunnerID}).Take(&runner).Error; err == nil &&
			runner.RunnerGroup.Type == app.RunnerGroupTypeInstall && runner.RunnerGroup.OwnerID != "" {
			grants[permissions.Object(orgID, permissions.KindInstall, runner.RunnerGroup.OwnerID)] = permissions.PermissionAll.ToStrPtr()
		}
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
			OrgID:       generics.NewNullString(orgID),
			Name:        app.PolicyNameRunnerJob,
			Permissions: grants,
		}},
	}
	if err := authz.New(authz.Params{DB: h.db}).EnsureManagedServiceAccountRole(roleCtx, acct.ID, role); err != nil {
		return fmt.Errorf("create job role: %w", err)
	}
	return nil
}
