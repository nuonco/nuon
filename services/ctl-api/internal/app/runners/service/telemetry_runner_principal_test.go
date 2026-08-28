package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const (
	telemetryTestOrgID     = "org-test"
	telemetryTestAppID     = "app-test"
	telemetryTestInstallID = "install-test"
	telemetryTestGroupID   = "runner-group-test"
	telemetryTestRunnerID  = "runner-test"
)

func setupTelemetryRunnerPrincipalDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	for _, statement := range []string{
		`CREATE TABLE runners (id TEXT PRIMARY KEY, org_id TEXT NOT NULL, status TEXT NOT NULL, runner_group_id TEXT NOT NULL, updated_at DATETIME, deleted_at INTEGER NOT NULL DEFAULT 0)`,
		`CREATE TABLE runner_groups (id TEXT PRIMARY KEY, org_id TEXT NOT NULL, owner_id TEXT NOT NULL, owner_type TEXT NOT NULL, type TEXT NOT NULL, updated_at DATETIME, deleted_at INTEGER NOT NULL DEFAULT 0)`,
		`CREATE TABLE installs (id TEXT PRIMARY KEY, org_id TEXT NOT NULL, app_id TEXT NOT NULL, deleted_at INTEGER NOT NULL DEFAULT 0)`,
		`INSERT INTO runner_groups (id, org_id, owner_id, owner_type, type) VALUES ('runner-group-test', 'org-test', 'install-test', 'installs', 'install')`,
		`INSERT INTO runners (id, org_id, status, runner_group_id) VALUES ('runner-test', 'org-test', 'active', 'runner-group-test')`,
		`INSERT INTO installs (id, org_id, app_id) VALUES ('install-test', 'org-test', 'app-test')`,
	} {
		require.NoError(t, db.Exec(statement).Error)
	}

	return db
}

func telemetryRunnerTestAccount() *app.Account {
	return &app.Account{
		AccountType: app.AccountTypeService,
		Subject:     telemetryTestRunnerID,
		OrgIDs:      []string{telemetryTestOrgID},
		Roles: []app.Role{{
			RoleType: app.RoleTypeRunner,
			OrgID:    generics.NewNullString(telemetryTestOrgID),
		}},
	}
}

func telemetryRunnerTestContext() context.Context {
	return cctx.SetOrgIDContext(context.Background(), telemetryTestOrgID)
}

func requireTelemetryRunnerAuthorizationError(t *testing.T, err error) {
	t.Helper()

	var authorizationErr stderr.ErrAuthorization
	require.ErrorAs(t, err, &authorizationErr)
	require.True(t, errors.Is(err, errTelemetryRunnerUnauthorized))
}

func TestResolveTelemetryRunnerPrincipal(t *testing.T) {
	svc := &service{db: setupTelemetryRunnerPrincipalDB(t)}

	principal, err := svc.resolveTelemetryRunnerPrincipal(telemetryRunnerTestContext(), telemetryRunnerTestAccount())
	require.NoError(t, err)
	require.Equal(t, telemetryRunnerPrincipal{
		OrgID:     telemetryTestOrgID,
		AppID:     telemetryTestAppID,
		InstallID: telemetryTestInstallID,
		RunnerID:  telemetryTestRunnerID,
	}, principal)
}

func TestResolveTelemetryRunnerPrincipalAllowsNonterminalRunnerStatuses(t *testing.T) {
	for _, status := range []app.RunnerStatus{
		app.RunnerStatusOffline,
		app.RunnerStatusDeprovisioning,
		app.RunnerStatusError,
		app.RunnerStatusUnknown,
	} {
		t.Run(status.String(), func(t *testing.T) {
			db := setupTelemetryRunnerPrincipalDB(t)
			require.NoError(t, db.Model(&app.Runner{}).
				Where(app.Runner{ID: telemetryTestRunnerID}).
				Update("status", status).Error)
			svc := &service{db: db}

			_, err := svc.resolveTelemetryRunnerPrincipal(telemetryRunnerTestContext(), telemetryRunnerTestAccount())
			require.NoError(t, err)
		})
	}
}

func TestResolveTelemetryRunnerPrincipalRejectsIneligibleRunnerStatuses(t *testing.T) {
	for _, status := range []app.RunnerStatus{app.RunnerStatusDisabled, app.RunnerStatusDeprovisioned} {
		t.Run(status.String(), func(t *testing.T) {
			db := setupTelemetryRunnerPrincipalDB(t)
			require.NoError(t, db.Model(&app.Runner{}).
				Where(app.Runner{ID: telemetryTestRunnerID}).
				Update("status", status).Error)
			svc := &service{db: db}

			_, err := svc.resolveTelemetryRunnerPrincipal(telemetryRunnerTestContext(), telemetryRunnerTestAccount())
			requireTelemetryRunnerAuthorizationError(t, err)
		})
	}
}

func TestResolveTelemetryRunnerPrincipalRejectsNonRunnerAccounts(t *testing.T) {
	tests := map[string]*app.Account{
		"user account": {
			AccountType: app.AccountTypeAuth0,
			Subject:     telemetryTestRunnerID,
			OrgIDs:      []string{telemetryTestOrgID},
			Roles: []app.Role{{
				RoleType: app.RoleTypeRunner,
				OrgID:    generics.NewNullString(telemetryTestOrgID),
			}},
		},
		"service account for another resource": {
			AccountType: app.AccountTypeService,
			Subject:     "install-stack-test",
			OrgIDs:      []string{telemetryTestOrgID},
			Roles: []app.Role{{
				RoleType: app.RoleTypeRunner,
				OrgID:    generics.NewNullString(telemetryTestOrgID),
			}},
		},
		"service account outside org": {
			AccountType: app.AccountTypeService,
			Subject:     telemetryTestRunnerID,
			OrgIDs:      []string{"org-other"},
			Roles: []app.Role{{
				RoleType: app.RoleTypeRunner,
				OrgID:    generics.NewNullString("org-other"),
			}},
		},
		"service account without runner role": {
			AccountType: app.AccountTypeService,
			Subject:     telemetryTestRunnerID,
			OrgIDs:      []string{telemetryTestOrgID},
			Roles: []app.Role{{
				RoleType: app.RoleTypeStack,
				OrgID:    generics.NewNullString(telemetryTestOrgID),
			}},
		},
	}

	for name, acct := range tests {
		t.Run(name, func(t *testing.T) {
			svc := &service{db: setupTelemetryRunnerPrincipalDB(t)}

			_, err := svc.resolveTelemetryRunnerPrincipal(telemetryRunnerTestContext(), acct)
			requireTelemetryRunnerAuthorizationError(t, err)
		})
	}
}

func TestResolveTelemetryRunnerPrincipalRejectsNonInstallRunner(t *testing.T) {
	db := setupTelemetryRunnerPrincipalDB(t)
	require.NoError(t, db.Model(&app.RunnerGroup{}).
		Where(app.RunnerGroup{ID: telemetryTestGroupID}).
		Updates(app.RunnerGroup{Type: app.RunnerGroupTypeOrg, OwnerType: "orgs"}).Error)
	svc := &service{db: db}

	_, err := svc.resolveTelemetryRunnerPrincipal(telemetryRunnerTestContext(), telemetryRunnerTestAccount())
	requireTelemetryRunnerAuthorizationError(t, err)
}
