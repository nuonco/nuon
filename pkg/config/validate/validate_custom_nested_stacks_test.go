package validate

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/publicnet"
	"github.com/nuonco/nuon/pkg/publicnet/templatefetch"
)

func useLocalTemplateClient(t *testing.T) {
	t.Helper()
	prev := templateOutputsClient
	templateOutputsClient = templatefetch.NewLocalClient()
	t.Cleanup(func() { templateOutputsClient = prev })
}

func TestValidateCustomNestedStackOutputs_NilStack(t *testing.T) {
	err := ValidateCustomNestedStackOutputs(&config.AppConfig{})
	assert.NoError(t, err)
}

func TestValidateCustomNestedStackOutputs_NoAdditionalStacks(t *testing.T) {
	err := ValidateCustomNestedStackOutputs(&config.AppConfig{
		Stack: &config.StackConfig{},
	})
	assert.NoError(t, err)
}

func TestValidateCustomNestedStackOutputs_NoCollisions(t *testing.T) {
	templateA := `
AWSTemplateFormatVersion: '2010-09-09'
Outputs:
  OutputA:
    Value: a
`
	templateB := `
AWSTemplateFormatVersion: '2010-09-09'
Outputs:
  OutputB:
    Value: b
`
	mux := http.NewServeMux()
	mux.HandleFunc("/a.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(templateA))
	})
	mux.HandleFunc("/b.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(templateB))
	})
	useLocalTemplateClient(t)
	server := httptest.NewServer(mux)
	defer server.Close()

	err := ValidateCustomNestedStackOutputs(&config.AppConfig{
		Stack: &config.StackConfig{
			CustomNestedStacks: []config.CustomNestedStack{
				{Name: "stack_a", TemplateURL: server.URL + "/a.yaml", Index: 0},
				{Name: "stack_b", TemplateURL: server.URL + "/b.yaml", Index: 1},
			},
		},
	})
	assert.NoError(t, err)
}

func TestValidateCustomNestedStackOutputs_CollisionBetweenAdditionalStacks(t *testing.T) {
	templateA := `
AWSTemplateFormatVersion: '2010-09-09'
Outputs:
  SharedOutput:
    Value: a
`
	templateB := `
AWSTemplateFormatVersion: '2010-09-09'
Outputs:
  SharedOutput:
    Value: b
`
	mux := http.NewServeMux()
	mux.HandleFunc("/a.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(templateA))
	})
	mux.HandleFunc("/b.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(templateB))
	})
	useLocalTemplateClient(t)
	server := httptest.NewServer(mux)
	defer server.Close()

	err := ValidateCustomNestedStackOutputs(&config.AppConfig{
		Stack: &config.StackConfig{
			CustomNestedStacks: []config.CustomNestedStack{
				{Name: "stack_a", TemplateURL: server.URL + "/a.yaml", Index: 0},
				{Name: "stack_b", TemplateURL: server.URL + "/b.yaml", Index: 1},
			},
		},
	})
	require.Error(t, err)
	assert.True(t, config.IsWarningErr(err))
	assert.Contains(t, err.Error(), "SharedOutput")
	assert.Contains(t, err.Error(), "stack_a")
	assert.Contains(t, err.Error(), "stack_b")
}

func TestValidateCustomNestedStackOutputs_CollisionWithFirstClass(t *testing.T) {
	vpcTemplate := `
AWSTemplateFormatVersion: '2010-09-09'
Outputs:
  VPC:
    Value: vpc-id
  RunnerSubnet:
    Value: subnet-id
`
	additionalTemplate := `
AWSTemplateFormatVersion: '2010-09-09'
Outputs:
  VPC:
    Value: colliding-vpc
`
	mux := http.NewServeMux()
	mux.HandleFunc("/vpc.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(vpcTemplate))
	})
	mux.HandleFunc("/additional.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(additionalTemplate))
	})
	useLocalTemplateClient(t)
	server := httptest.NewServer(mux)
	defer server.Close()

	err := ValidateCustomNestedStackOutputs(&config.AppConfig{
		Stack: &config.StackConfig{
			VPCNestedTemplateURL: server.URL + "/vpc.yaml",
			CustomNestedStacks: []config.CustomNestedStack{
				{Name: "my_stack", TemplateURL: server.URL + "/additional.yaml", Index: 0},
			},
		},
	})
	require.Error(t, err)
	assert.True(t, config.IsWarningErr(err))
	assert.Contains(t, err.Error(), "VPC")
	assert.Contains(t, err.Error(), "first-class output will take precedence")
}

func TestValidateCustomNestedStackOutputs_UnfetchableTemplateSkipped(t *testing.T) {
	useLocalTemplateClient(t)
	err := ValidateCustomNestedStackOutputs(&config.AppConfig{
		Stack: &config.StackConfig{
			CustomNestedStacks: []config.CustomNestedStack{
				{Name: "my_stack", TemplateURL: "http://invalid.localhost.test/stack.yaml", Index: 0},
			},
		},
	})
	assert.NoError(t, err)
}

func TestValidateTemplateURLsRefuseNonPublicHosts(t *testing.T) {
	azureWithRunner := func(url string) *config.AppConfig {
		return &config.AppConfig{
			Stack: &config.StackConfig{Type: "azure-bicep", RunnerNestedTemplateURL: url},
			Permissions: &config.PermissionsConfig{
				ProvisionRole: &config.AppAWSIAMRole{CloudPlatform: "azure"},
			},
		}
	}

	tests := map[string]struct {
		validate func(*config.AppConfig) error
		cfg      func(url string) *config.AppConfig
	}{
		"custom nested stack": {
			validate: ValidateCustomNestedStackOutputs,
			cfg: func(url string) *config.AppConfig {
				return &config.AppConfig{Stack: &config.StackConfig{
					CustomNestedStacks: []config.CustomNestedStack{{Name: "my_stack", TemplateURL: url}},
				}}
			},
		},
		"vpc nested template": {
			validate: ValidateCustomNestedStackOutputs,
			cfg: func(url string) *config.AppConfig {
				return &config.AppConfig{Stack: &config.StackConfig{
					VPCNestedTemplateURL: url,
					CustomNestedStacks:   []config.CustomNestedStack{{Name: "my_stack"}},
				}}
			},
		},
		"azure runner nested template": {
			validate: ValidateAzureRunnerIdentities,
			cfg:      azureWithRunner,
		},
	}

	for name, tt := range tests {
		for _, url := range []string{
			"https://127.0.0.1/template.yaml",
			"https://169.254.169.254/latest/meta-data/",
		} {
			t.Run(name+" "+url, func(t *testing.T) {
				err := tt.validate(tt.cfg(url))
				require.ErrorIs(t, err, publicnet.ErrNonPublic)
				assert.False(t, config.IsWarningErr(err))
			})
		}
	}
}

func TestValidateAzureCustomNestedStacksChecksInstallOverrides(t *testing.T) {
	err := validateAzureCustomNestedStacks(&config.AppConfig{
		Stack: &config.StackConfig{Type: "azure-bicep"},
		Installs: []*config.Install{
			{
				Name: "example",
				StackOverrides: &config.InstallStackOverrides{
					CustomNestedStacks: []config.CustomNestedStack{
						{Name: "storage", Index: 0, TemplateURL: "./arm/storage.bicep", Contents: "param location string"},
					},
				},
			},
		},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "az bicep build --file ./arm/storage.bicep --outfile ./arm/storage.json")
}
