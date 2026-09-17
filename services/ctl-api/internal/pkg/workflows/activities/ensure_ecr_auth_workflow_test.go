package activities

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"

	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	"github.com/nuonco/nuon/pkg/plugins/configs"
)

func ecrSandboxCfg() *configs.OCIRegistryRepository {
	return &configs.OCIRegistryRepository{
		Plugin:       "oci",
		RegistryType: configs.OCIRegistryTypeECR,
		Repository:   ecrRepositoryURI,
		Region:       "us-west-2",
		ECRAuth: &awscredentials.Config{
			Region: "us-west-2",
			AssumeRole: &awscredentials.AssumeRoleConfig{
				RoleARN:     "arn:aws:iam::111122223333:role/ctl-api-orgs-account-iam-access",
				SessionName: "sandbox-build",
			},
		},
	}
}

func runEnsureECRAuth(t *testing.T, env *testsuite.TestWorkflowEnvironment, cfg *configs.OCIRegistryRepository) *configs.OCIRegistryRepository {
	t.Helper()
	env.ExecuteWorkflow(func(ctx workflow.Context) (*configs.OCIRegistryRepository, error) {
		if err := EnsureECRAuth(ctx, cfg); err != nil {
			return nil, err
		}
		return cfg, nil
	})
	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	var out *configs.OCIRegistryRepository
	require.NoError(t, env.GetWorkflowResult(&out))
	return out
}

func TestEnsureECRAuthEmbedsCredentials(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.OnActivity((*Activities).GetECRAccessToken, mock.Anything, mock.Anything, mock.Anything).
		Return(&ECRAccessToken{
			Username:      "AWS",
			Password:      "minted-token",
			ServerAddress: "111122223333.dkr.ecr.us-west-2.amazonaws.com",
		}, nil).Once()

	got := runEnsureECRAuth(t, env, ecrSandboxCfg())

	require.Equal(t, configs.OCIRegistryTypePrivateOCI, got.RegistryType)
	require.Nil(t, got.ECRAuth, "the runner must not be able to fall back to assume-role")
	require.Equal(t, "AWS", got.OCIAuth.Username)
	require.Equal(t, "minted-token", got.OCIAuth.Password)
	require.Equal(t, "111122223333.dkr.ecr.us-west-2.amazonaws.com", got.LoginServer)
	env.AssertExpectations(t)
}

func TestEnsureECRAuthLeavesOtherRegistriesAlone(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()

	cfg := &configs.OCIRegistryRepository{RegistryType: configs.OCIRegistryTypeGAR, Repository: "gar/repo"}
	got := runEnsureECRAuth(t, env, cfg)

	require.Equal(t, configs.OCIRegistryTypeGAR, got.RegistryType)
	require.Nil(t, got.OCIAuth)
}
