package activities

import (
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/workflow"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/plugins/configs"
)

func TestImagePullAuthenticationVersioning(t *testing.T) {
	for name, tc := range map[string]struct {
		version      workflow.Version
		registry     configs.OCIRegistryType
		registration *configs.ACRAppRegistration
		activity     string
	}{
		"replay legacy GAR activity":                    {version: workflow.DefaultVersion, registry: configs.OCIRegistryTypeGAR, activity: "gar"},
		"new GAR execution defers to job authorization": {version: 1, registry: configs.OCIRegistryTypeGAR},
		"new ECR execution defers to job authorization": {version: 1, registry: configs.OCIRegistryTypeECR},
		"ACR ambient unchanged":                         {version: 1, registry: configs.OCIRegistryTypeACR},
		"ACR secret unchanged":                          {version: 1, registry: configs.OCIRegistryTypeACR, registration: &configs.ACRAppRegistration{ComponentID: "component", TenantID: "tenant", ClientID: "client", ClientSecretName: "secret"}, activity: "acr"},
		"ACR certificate unchanged":                     {version: 1, registry: configs.OCIRegistryTypeACR, registration: &configs.ACRAppRegistration{ComponentID: "component", TenantID: "tenant", ClientID: "client", ClientCertificateName: "cert"}, activity: "acr"},
	} {
		t.Run(name, func(t *testing.T) {
			var suite testsuite.WorkflowTestSuite
			env := suite.NewTestWorkflowEnvironment()
			env.OnGetVersion("image-pull-auth-cloud-connection", workflow.DefaultVersion, 1).Return(tc.version)
			if tc.activity == "gar" {
				env.OnActivity((&Activities{}).GetGARAccessToken, mock.Anything, &GetGARAccessTokenRequest{ServiceAccountEmail: "acme@example.com", WorkloadIdentityProvider: "provider"}).Return(&GARAccessToken{Username: "oauth2accesstoken", Password: "acme-token"}, nil).Once()
			}
			if tc.activity == "acr" {
				env.OnActivity((&Activities{}).GetACRAccessToken, mock.Anything, mock.Anything).Return(&ACRAccessToken{Username: "acme", Password: "acme-token"}, nil).Once()
			}
			env.ExecuteWorkflow(func(ctx workflow.Context) (*configs.OCIRegistryRepository, error) {
				cfg := &configs.OCIRegistryRepository{RegistryType: tc.registry, ServiceAccountEmail: "acme@example.com", WorkloadIdentityProvider: "provider", ACRAppRegistration: tc.registration}
				plan := &plantypes.BuildPlan{ContainerImagePullPlan: &plantypes.ContainerImagePullPlan{RepoCfg: cfg}}
				if err := EnsureContainerImagePullAuth(ctx, plan); err != nil {
					return nil, err
				}
				return cfg, nil
			})
			require.NoError(t, env.GetWorkflowError())
			var cfg configs.OCIRegistryRepository
			require.NoError(t, env.GetWorkflowResult(&cfg))
			if tc.activity != "" {
				require.Equal(t, "acme-token", cfg.OCIAuth.Password)
			} else {
				require.Nil(t, cfg.OCIAuth)
			}
			env.AssertExpectations(t)
		})
	}
}
