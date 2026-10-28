package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/keys"
	dbgenerics "github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

func (s *InstallsServiceTestSuite) TestMCPStackSetupOmitsSecretsForReadOnly() {
	install := s.createTestInstall()
	require.NotNil(s.T(), install.InstallStack)

	version := s.deps.Seeder.CreateInstallStackVersion(
		s.ctx, s.T(), install.ID, install.InstallStack.ID, install.AppConfigID,
	)

	envelope, err := json.Marshal(map[string]string{
		"inputs_tfvars":   "region = \"us-west-2\"\napi_token = \"secret-input\"\n",
		"provider_tfvars": "region = \"us-west-2\"\n",
		"secrets_tfvars":  "db_password = \"super-secret\"\n",
	})
	require.NoError(s.T(), err)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Model(&app.InstallStackVersion{}).
		Where("id = ?", version.ID).
		Update("terraform_contents", envelope).Error)

	run := &app.InstallStackVersionRun{
		InstallStackVersionID: version.ID,
		Data: dbgenerics.ToHstore(map[string]string{
			"runner_token": "plaintext-runner-token",
		}),
		RunType: app.StackVersionRunTypeOutOfBand,
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(run).Error)

	writeCtx := cctx.SetAccountContext(context.Background(), &app.Account{
		ID: s.testAcc.ID,
		AllPermissions: permissions.Set{
			s.testOrg.ID: permissions.PermissionRead,
		},
	})
	writeCtx = context.WithValue(writeCtx, keys.OrgIDCtxKey, s.testOrg.ID)

	writeSetup := s.installsService.mcpStackSetup(writeCtx, s.testOrg.ID, version.ID)
	require.NotNil(s.T(), writeSetup)
	assert.Contains(s.T(), writeSetup.SecretsTfvars, "super-secret")
	assert.Contains(s.T(), writeSetup.InputsTfvars, "secret-input")
	assert.Contains(s.T(), writeSetup.ProviderTfvars, "us-west-2")
	require.NotEmpty(s.T(), writeSetup.Outputs)
	assert.Equal(s.T(), "plaintext-runner-token", writeSetup.Outputs["runner_token"])

	readOnlyCtx := keys.WithTokenRole(writeCtx, string(app.RoleTypeOrgReadOnly))
	readSetup := s.installsService.mcpStackSetup(readOnlyCtx, s.testOrg.ID, version.ID)
	require.NotNil(s.T(), readSetup)
	assert.Empty(s.T(), readSetup.SecretsTfvars)
	assert.Empty(s.T(), readSetup.InputsTfvars)
	assert.Nil(s.T(), readSetup.Outputs)
	assert.Contains(s.T(), readSetup.ProviderTfvars, "us-west-2")
	assert.NotEmpty(s.T(), readSetup.Cloud)
}

func TestRedactStackSetupSecretsForReadOnly(t *testing.T) {
	t.Parallel()

	setup := &mcpStackSetup{
		InputsTfvars:   "secret-input",
		SecretsTfvars:  "super-secret",
		ProviderTfvars: "region",
		Outputs: map[string]any{
			"token": "value",
		},
	}

	ctx := keys.WithTokenRole(context.Background(), string(app.RoleTypeOrgReadOnly))
	redactStackSetupSecretsForReadOnly(ctx, setup)

	assert.Empty(t, setup.InputsTfvars)
	assert.Empty(t, setup.SecretsTfvars)
	assert.Nil(t, setup.Outputs)
	assert.Equal(t, "region", setup.ProviderTfvars)

	kept := &mcpStackSetup{
		InputsTfvars:  "keep-me",
		SecretsTfvars: "keep-secret",
		Outputs:       map[string]any{"k": "v"},
	}
	redactStackSetupSecretsForReadOnly(context.Background(), kept)
	assert.Equal(t, "keep-me", kept.InputsTfvars)
	assert.Equal(t, "keep-secret", kept.SecretsTfvars)
	assert.Equal(t, map[string]any{"k": "v"}, kept.Outputs)
}
