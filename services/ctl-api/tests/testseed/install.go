package testseed

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	fakegenerics "github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

// BuildInstall creates an app.Install with fake defaults for the given app.
// The returned install is in-memory only and does not set config FK fields
// (those require DB queries — use CreateInstall for a fully-populated install).
func BuildInstall(a *app.App) *app.Install {
	acct := BuildAccount()

	return &app.Install{
		ID:          domains.NewInstallID(),
		Name:        fmt.Sprintf("install-%s", fakegenerics.GetFakeObj[string]()),
		OrgID:       a.OrgID,
		CreatedBy:   *acct,
		CreatedByID: acct.ID,
		AppID:       a.ID,
		AWSAccount: &app.AWSAccount{
			Region: "us-west-2",
		},
	}
}

// CreateInstall builds and persists an install to the database for the given app.
// Requires that CreateAppConfig has already been called for the app — it queries for
// the latest AppSandboxConfig, AppRunnerConfig, and AppConfig to set FK fields.
// Creates associated AWSAccount, InstallSandbox, and InstallStack inline.
// Uses org/account from context if available.
func (s *Seeder) CreateInstall(ctx context.Context, t *testing.T, a *app.App) *app.Install {
	i := BuildInstall(a)
	if orgID, err := cctx.OrgIDFromContext(ctx); err == nil {
		i.OrgID = orgID
	}
	if accountID, err := cctx.AccountIDFromContext(ctx); err == nil {
		i.CreatedBy = app.Account{}
		i.CreatedByID = accountID
	}

	// Load latest configs from the app (created by CreateAppConfig).
	var sandboxCfg app.AppSandboxConfig
	require.NoError(t, s.db.WithContext(ctx).
		Where("app_id = ?", a.ID).Order("created_at DESC").First(&sandboxCfg).Error,
		"CreateInstall requires CreateAppConfig to have been called first")
	i.AppSandboxConfigID = sandboxCfg.ID

	var runnerCfg app.AppRunnerConfig
	require.NoError(t, s.db.WithContext(ctx).
		Where("app_id = ?", a.ID).Order("created_at DESC").First(&runnerCfg).Error)
	i.AppRunnerConfigID = runnerCfg.ID

	var appCfg app.AppConfig
	require.NoError(t, s.db.WithContext(ctx).
		Where("app_id = ?", a.ID).Order("created_at DESC").First(&appCfg).Error)
	i.AppConfigID = appCfg.ID

	// Create associated objects inline (mirrors helpers.CreateInstall).
	i.InstallSandbox = app.InstallSandbox{
		Status: app.InstallSandboxStatusQueued,
		TerraformWorkspace: app.TerraformWorkspace{
			ID: domains.NewTerraformWorkspaceID(),
		},
	}
	i.InstallStack = &app.InstallStack{
		InstallStackOutputs: app.InstallStackOutputs{
			Data: dbgenerics.ToHstore(map[string]string{}),
		},
	}

	res := s.db.WithContext(ctx).Create(i)
	require.NoError(t, res.Error)
	return i
}

// CreateInstallComponent persists an InstallComponent linking an install to a component.
// OrgID and CreatedByID are populated by the BeforeCreate hook from context.
func (s *Seeder) CreateInstallComponent(ctx context.Context, t *testing.T, installID, componentID string) *app.InstallComponent {
	ic := &app.InstallComponent{
		InstallID:   installID,
		ComponentID: componentID,
		Status:      app.InstallComponentStatusPending,
	}
	res := s.db.WithContext(ctx).Create(ic)
	require.NoError(t, res.Error)
	return ic
}
