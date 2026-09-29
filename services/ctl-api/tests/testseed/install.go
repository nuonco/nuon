package testseed

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"

	fakegenerics "github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

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

func (s *Seeder) CreateInstall(ctx context.Context, t *testing.T, a *app.App) *app.Install {
	i := BuildInstall(a)
	if orgID, err := cctx.OrgIDFromContext(ctx); err == nil {
		i.OrgID = orgID
	}
	if accountID, err := cctx.AccountIDFromContext(ctx); err == nil {
		i.CreatedBy = app.Account{}
		i.CreatedByID = accountID
	}

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

func (s *Seeder) CreateInstallDeploy(ctx context.Context, t *testing.T, installComponentID, componentBuildID string) *app.InstallDeploy {
	deploy := &app.InstallDeploy{
		InstallComponentID: installComponentID,
		ComponentBuildID:   componentBuildID,
		Status:             app.InstallDeployStatusActive,
		StatusV2:           app.NewCompositeStatus(ctx, app.StatusSuccess),
		Type:               app.InstallDeployTypeApply,
	}
	res := s.db.WithContext(ctx).Create(deploy)
	require.NoError(t, res.Error)
	return deploy
}

func (s *Seeder) CreateInstallInputs(ctx context.Context, t *testing.T, installID, appInputConfigID string, values map[string]*string) *app.InstallInputs {
	hstore := pgtype.Hstore{}
	for k, v := range values {
		hstore[k] = v
	}
	inputs := &app.InstallInputs{
		InstallID:        installID,
		AppInputConfigID: appInputConfigID,
		Values:           hstore,
	}
	res := s.db.WithContext(ctx).Create(inputs)
	require.NoError(t, res.Error)
	return inputs
}

func (s *Seeder) CreateInstallStackVersion(ctx context.Context, t *testing.T, installID, installStackID, appConfigID string) *app.InstallStackVersion {
	sv := &app.InstallStackVersion{
		InstallID:      installID,
		InstallStackID: installStackID,
		AppConfigID:    appConfigID,
		PhoneHomeID:    domains.NewInstallStackVersionID(),
		Status:         app.NewCompositeStatus(ctx, app.StatusPending),
		Checksum:       "test-checksum",
		Contents:       []byte("{}"),
	}
	res := s.db.WithContext(ctx).Create(sv)
	require.NoError(t, res.Error)
	return sv
}
