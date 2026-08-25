package plan

import (
	"testing"

	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	gcpcredentials "github.com/nuonco/nuon/pkg/gcp/credentials"
	"github.com/nuonco/nuon/pkg/plugins/configs"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageNameSegment(t *testing.T) {
	cases := map[string]string{
		"img_nuon_ctl_api":                 "img-nuon-ctl-api",
		"img_altinity_clickhouse_operator": "img-altinity-clickhouse-operator",
		"My Component":                     "my-component",
		"foo   bar":                        "foo-bar",
		"foo--bar":                         "foo-bar",
		"-leading-and-trailing-":           "leading-and-trailing",
		"trailing___":                      "trailing",
		"...dots...":                       "dots",
		"CTL API (v2)":                     "ctl-api-v2",
		"":                                 "app",
		"___":                              "app",
	}

	for in, want := range cases {
		assert.Equal(t, want, imageNameSegment(in), "input %q", in)
	}
}

func sandboxOutputs(outputs map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"sandbox": map[string]interface{}{
			"outputs": outputs,
		},
	}
}

func TestInstallRegistryLoginServer(t *testing.T) {
	tests := []struct {
		name      string
		stack     *app.InstallStack
		stateData map[string]interface{}
		want      string
	}{
		{
			name: "aws reads the ecr registry url",
			stack: &app.InstallStack{InstallStackOutputs: app.InstallStackOutputs{
				AWSStackOutputs: &app.AWSStackOutputs{Region: "us-west-2"},
			}},
			stateData: sandboxOutputs(map[string]interface{}{
				"ecr": map[string]interface{}{"registry_url": "123456789012.dkr.ecr.us-west-2.amazonaws.com"},
			}),
			want: "123456789012.dkr.ecr.us-west-2.amazonaws.com",
		},
		{
			name: "https prefix is trimmed",
			stack: &app.InstallStack{InstallStackOutputs: app.InstallStackOutputs{
				AWSStackOutputs: &app.AWSStackOutputs{Region: "us-west-2"},
			}},
			stateData: sandboxOutputs(map[string]interface{}{
				"ecr": map[string]interface{}{"registry_url": "https://123456789012.dkr.ecr.us-west-2.amazonaws.com"},
			}),
			want: "123456789012.dkr.ecr.us-west-2.amazonaws.com",
		},
		{
			name: "azure reads the acr login server",
			stack: &app.InstallStack{InstallStackOutputs: app.InstallStackOutputs{
				AzureStackOutputs: &app.AzureStackOutputs{},
			}},
			stateData: sandboxOutputs(map[string]interface{}{
				"acr": map[string]interface{}{"login_server": "acme.azurecr.io"},
			}),
			want: "acme.azurecr.io",
		},
		{
			name: "gcp reads the gar registry url",
			stack: &app.InstallStack{InstallStackOutputs: app.InstallStackOutputs{
				GCPStackOutputs: &app.GCPStackOutputs{Region: "us-central1"},
			}},
			stateData: sandboxOutputs(map[string]interface{}{
				"gar": map[string]interface{}{"registry_url": "us-central1-docker.pkg.dev"},
			}),
			want: "us-central1-docker.pkg.dev",
		},
		{
			name:      "no cloud outputs yields no login server",
			stack:     &app.InstallStack{InstallStackOutputs: app.InstallStackOutputs{}},
			stateData: sandboxOutputs(map[string]interface{}{}),
			want:      "",
		},
		{
			// A sandbox with no registry output is valid: the install can still
			// run actions that pull public images.
			name: "missing registry output yields no login server",
			stack: &app.InstallStack{InstallStackOutputs: app.InstallStackOutputs{
				AWSStackOutputs: &app.AWSStackOutputs{Region: "us-west-2"},
			}},
			stateData: sandboxOutputs(map[string]interface{}{}),
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, installRegistryLoginServer(tt.stateData, tt.stack))
		})
	}
}

func TestGetInstallRegistryPullConfig(t *testing.T) {
	awsAuth := &awscredentials.Config{Region: "us-west-2"}

	t.Run("aws uses the role-selected ecr credentials", func(t *testing.T) {
		stack := &app.InstallStack{InstallStackOutputs: app.InstallStackOutputs{
			AWSStackOutputs: &app.AWSStackOutputs{Region: "us-west-2"},
		}}

		cfg := getInstallRegistryPullConfig(
			"123456789012.dkr.ecr.us-west-2.amazonaws.com/org/app/tools",
			"123456789012.dkr.ecr.us-west-2.amazonaws.com",
			stack,
			&CloudAuth{AWS: awsAuth},
		)

		require.NotNil(t, cfg)
		assert.Equal(t, configs.OCIRegistryTypeECR, cfg.RegistryType)
		assert.Equal(t, "us-west-2", cfg.Region)
		assert.Equal(t, awsAuth, cfg.ECRAuth)
		// ECR trims the server address off the repository, so it has to carry it.
		assert.Equal(t, "123456789012.dkr.ecr.us-west-2.amazonaws.com/org/app/tools", cfg.Repository)
	})

	t.Run("azure falls back to the runner's ambient identity", func(t *testing.T) {
		stack := &app.InstallStack{InstallStackOutputs: app.InstallStackOutputs{
			AzureStackOutputs: &app.AzureStackOutputs{},
		}}

		cfg := getInstallRegistryPullConfig("acme.azurecr.io/org/app/tools", "acme.azurecr.io", stack, &CloudAuth{})

		require.NotNil(t, cfg)
		assert.Equal(t, configs.OCIRegistryTypeACR, cfg.RegistryType)
		require.NotNil(t, cfg.ACRAuth)
		assert.True(t, cfg.ACRAuth.UseDefault)
	})

	t.Run("gcp impersonates the selected service account", func(t *testing.T) {
		stack := &app.InstallStack{InstallStackOutputs: app.InstallStackOutputs{
			GCPStackOutputs: &app.GCPStackOutputs{Region: "us-central1"},
		}}

		cfg := getInstallRegistryPullConfig(
			"us-central1-docker.pkg.dev/proj/repo/tools",
			"us-central1-docker.pkg.dev",
			stack,
			&CloudAuth{GCP: &gcpcredentials.Config{ImpersonateServiceAccount: "svc@proj.iam.gserviceaccount.com"}},
		)

		require.NotNil(t, cfg)
		assert.Equal(t, configs.OCIRegistryTypeGAR, cfg.RegistryType)
		assert.Equal(t, "us-central1", cfg.Region)
		assert.Equal(t, "svc@proj.iam.gserviceaccount.com", cfg.ServiceAccountEmail)
	})

	t.Run("unknown cloud yields no config", func(t *testing.T) {
		stack := &app.InstallStack{InstallStackOutputs: app.InstallStackOutputs{}}

		assert.Nil(t, getInstallRegistryPullConfig("example.com/tools", "example.com", stack, &CloudAuth{}))
	})
}
