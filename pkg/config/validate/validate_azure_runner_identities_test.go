package validate

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
)

func azureRunnerTestServer(t *testing.T, body string) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv.URL + "/runner.json"
}

func azureAppCfg(url string) *config.AppConfig {
	return &config.AppConfig{
		Stack: &config.StackConfig{Type: "azure-bicep", RunnerNestedTemplateURL: url},
		Permissions: &config.PermissionsConfig{
			ProvisionRole: &config.AppAWSIAMRole{CloudPlatform: "azure"},
		},
	}
}

func TestValidateAzureRunnerIdentities_MissingParam(t *testing.T) {
	url := azureRunnerTestServer(t, `{"parameters":{"nuonInstallID":{"type":"string"}},"resources":[]}`)
	err := ValidateAzureRunnerIdentities(azureAppCfg(url))
	require.Error(t, err)
	require.Contains(t, err.Error(), "userAssignedIdentities")
}

func TestValidateAzureRunnerIdentities_ParamPresent(t *testing.T) {
	url := azureRunnerTestServer(t, `{"parameters":{"userAssignedIdentities":{"type":"object"}},"resources":[]}`)
	require.NoError(t, ValidateAzureRunnerIdentities(azureAppCfg(url)))
}

func TestValidateAzureRunnerIdentities_BuiltInRunnerSkips(t *testing.T) {
	cfg := azureAppCfg("")
	cfg.Stack.RunnerNestedTemplateURL = ""
	require.NoError(t, ValidateAzureRunnerIdentities(cfg))
}

func TestValidateAzureRunnerIdentities_NonAzureRoleSkips(t *testing.T) {
	url := azureRunnerTestServer(t, `{"parameters":{},"resources":[]}`)
	cfg := azureAppCfg(url)
	cfg.Permissions.ProvisionRole.CloudPlatform = "aws"
	require.NoError(t, ValidateAzureRunnerIdentities(cfg))
}
