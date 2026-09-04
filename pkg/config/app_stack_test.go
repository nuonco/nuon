package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStackConfig_Parse_NoCustomNestedStacks(t *testing.T) {
	cfg := &StackConfig{
		Type:                    "aws-cloudformation",
		Name:                    "my-stack",
		Description:             "test stack",
		VPCNestedTemplateURL:    "https://s3.amazonaws.com/bucket/vpc.yaml",
		RunnerNestedTemplateURL: "https://s3.amazonaws.com/bucket/runner.yaml",
	}
	require.NoError(t, cfg.parse())
}

func TestStackConfig_Parse_EmptyCustomNestedStacks(t *testing.T) {
	cfg := &StackConfig{
		Type:                    "aws-cloudformation",
		Name:                    "my-stack",
		Description:             "test stack",
		VPCNestedTemplateURL:    "https://s3.amazonaws.com/bucket/vpc.yaml",
		RunnerNestedTemplateURL: "https://s3.amazonaws.com/bucket/runner.yaml",
		CustomNestedStacks:      []CustomNestedStack{},
	}
	require.NoError(t, cfg.parse())
}

func TestStackConfig_Parse_ValidCustomNestedStacks(t *testing.T) {
	cfg := &StackConfig{
		Type:                    "aws-cloudformation",
		Name:                    "my-stack",
		Description:             "test stack",
		VPCNestedTemplateURL:    "https://s3.amazonaws.com/bucket/vpc.yaml",
		RunnerNestedTemplateURL: "https://s3.amazonaws.com/bucket/runner.yaml",
		CustomNestedStacks: []CustomNestedStack{
			{Name: "k8s_namespaces", TemplateURL: "https://s3.amazonaws.com/bucket/template.yaml", Index: 0},
		},
	}
	require.NoError(t, cfg.parse())
}

func TestStackConfig_Parse_DuplicateIndex(t *testing.T) {
	cfg := &StackConfig{
		Type:                    "aws-cloudformation",
		Name:                    "my-stack",
		Description:             "test stack",
		VPCNestedTemplateURL:    "https://s3.amazonaws.com/bucket/vpc.yaml",
		RunnerNestedTemplateURL: "https://s3.amazonaws.com/bucket/runner.yaml",
		CustomNestedStacks: []CustomNestedStack{
			{Name: "preview_bucket", TemplateURL: "./cloudformation/s3-bucket/stack.yaml", Index: 0},
			{Name: "rds_subnet", TemplateURL: "./cloudformation/rds-subnet/stack.yaml", Index: 0},
		},
	}
	err := cfg.parse()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unique index")
}

func TestStackConfig_Parse_UniqueIndices(t *testing.T) {
	cfg := &StackConfig{
		Type:                    "aws-cloudformation",
		Name:                    "my-stack",
		Description:             "test stack",
		VPCNestedTemplateURL:    "https://s3.amazonaws.com/bucket/vpc.yaml",
		RunnerNestedTemplateURL: "https://s3.amazonaws.com/bucket/runner.yaml",
		CustomNestedStacks: []CustomNestedStack{
			{Name: "preview_bucket", TemplateURL: "./cloudformation/s3-bucket/stack.yaml", Index: 0},
			{Name: "rds_subnet", TemplateURL: "./cloudformation/rds-subnet/stack.yaml", Index: 1},
		},
	}
	require.NoError(t, cfg.parse())
}

func TestStackConfig_Parse_DeploymentScope(t *testing.T) {
	for name, tc := range map[string]struct {
		cfg     StackConfig
		wantErr string
	}{
		"azure subscription": {
			cfg: StackConfig{Type: "azure-bicep", Name: "s", Description: "d", DeploymentScope: StackDeploymentScopeSubscription},
		},
		"azure resource group": {
			cfg: StackConfig{Type: "azure-bicep", Name: "s", Description: "d", DeploymentScope: StackDeploymentScopeResourceGroup},
		},
		"azure unset": {
			cfg: StackConfig{Type: "azure-bicep", Name: "s", Description: "d"},
		},
		"gcp unset": {
			cfg: StackConfig{Type: "gcp-terraform", Name: "s", Description: "d"},
		},
		"aws subscription rejected": {
			cfg: StackConfig{
				Type:                    "aws-cloudformation",
				Name:                    "s",
				Description:             "d",
				VPCNestedTemplateURL:    "https://s3.amazonaws.com/bucket/vpc.yaml",
				RunnerNestedTemplateURL: "https://s3.amazonaws.com/bucket/runner.yaml",
				DeploymentScope:         StackDeploymentScopeSubscription,
			},
			wantErr: "only supported when type is azure-bicep",
		},
		"unknown scope rejected": {
			cfg:     StackConfig{Type: "azure-bicep", Name: "s", Description: "d", DeploymentScope: "tenant"},
			wantErr: `deployment_scope must be "resource_group" or "subscription"`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			err := tc.cfg.parse()
			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestStackConfig_Parse_AWSCloudFormation_MissingVPCTemplateURL(t *testing.T) {
	cfg := &StackConfig{
		Type:                    "aws-cloudformation",
		Name:                    "my-stack",
		Description:             "test stack",
		RunnerNestedTemplateURL: "https://s3.amazonaws.com/bucket/runner.yaml",
	}
	err := cfg.parse()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "vpc_nested_template_url is required when type is aws-cloudformation")
}

func TestStackConfig_Parse_AWSCloudFormation_MissingRunnerTemplateURL(t *testing.T) {
	cfg := &StackConfig{
		Type:                 "aws-cloudformation",
		Name:                 "my-stack",
		Description:          "test stack",
		VPCNestedTemplateURL: "https://s3.amazonaws.com/bucket/vpc.yaml",
	}
	err := cfg.parse()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "runner_nested_template_url is required when type is aws-cloudformation")
}

func TestStackConfig_Parse_MissingName(t *testing.T) {
	cfg := &StackConfig{
		CustomNestedStacks: []CustomNestedStack{
			{TemplateURL: "https://s3.amazonaws.com/bucket/template.yaml", Index: 0},
		},
	}
	err := cfg.parse()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestStackConfig_Parse_MissingTemplateURL(t *testing.T) {
	cfg := &StackConfig{
		CustomNestedStacks: []CustomNestedStack{
			{Name: "my_stack", Index: 0},
		},
	}
	err := cfg.parse()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "template_url is required")
}

func TestStackConfig_Parse_NonS3URLAccepted(t *testing.T) {
	// Custom nested stack template URLs no longer require S3 — go-getter handles fetching.
	cfg := &StackConfig{
		Type:                    "aws-cloudformation",
		Name:                    "my-stack",
		Description:             "test stack",
		VPCNestedTemplateURL:    "https://s3.amazonaws.com/bucket/vpc.yaml",
		RunnerNestedTemplateURL: "https://s3.amazonaws.com/bucket/runner.yaml",
		CustomNestedStacks: []CustomNestedStack{
			{Name: "my_stack", TemplateURL: "https://example.com/template.yaml", Index: 0},
		},
	}
	require.NoError(t, cfg.parse())
}

func TestStackConfig_Parse_FilePathURLAccepted(t *testing.T) {
	cfg := &StackConfig{
		Type:                    "aws-cloudformation",
		Name:                    "my-stack",
		Description:             "test stack",
		VPCNestedTemplateURL:    "https://s3.amazonaws.com/bucket/vpc.yaml",
		RunnerNestedTemplateURL: "https://s3.amazonaws.com/bucket/runner.yaml",
		CustomNestedStacks: []CustomNestedStack{
			{Name: "my_stack", TemplateURL: "./templates/custom.yaml", Index: 0},
		},
	}
	require.NoError(t, cfg.parse())
}

func TestStackConfig_Parse_VPCTemplateURLValidation(t *testing.T) {
	cfg := &StackConfig{
		VPCNestedTemplateURL: "https://example.com/vpc.yaml",
	}
	err := cfg.parse()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not an S3 URL")
}

func TestStackConfig_Parse_RunnerTemplateURLValidation(t *testing.T) {
	cfg := &StackConfig{
		RunnerNestedTemplateURL: "https://example.com/runner.yaml",
	}
	err := cfg.parse()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not an S3 URL")
}

func TestStackConfig_Parse_ValidFirstClassS3URLs(t *testing.T) {
	cfg := &StackConfig{
		Type:                    "aws-cloudformation",
		Name:                    "my-stack",
		Description:             "test stack",
		VPCNestedTemplateURL:    "https://s3.amazonaws.com/bucket/vpc.yaml",
		RunnerNestedTemplateURL: "https://nuon-artifacts.s3.us-west-2.amazonaws.com/runner.yaml",
	}
	require.NoError(t, cfg.parse())
}

func TestStackConfig_Parse_ValidParameters(t *testing.T) {
	cfg := &StackConfig{
		Type:                    "aws-cloudformation",
		Name:                    "my-stack",
		Description:             "test stack",
		VPCNestedTemplateURL:    "https://s3.amazonaws.com/bucket/vpc.yaml",
		RunnerNestedTemplateURL: "https://s3.amazonaws.com/bucket/runner.yaml",
		CustomNestedStacks: []CustomNestedStack{
			{
				Name:        "k8s_namespaces",
				TemplateURL: "https://s3.amazonaws.com/bucket/template.yaml",
				Index:       0,
				Parameters: map[string]string{
					"Namespaces":  "{{.nuon.install.inputs.namespaces}}",
					"ClusterName": "{{ .nuon.install.inputs.cluster_name }}",
					// Parameter values are full templates: conditionals, sprig
					// pipelines, composition with literals, and bare literals.
					"RootDomain":  "{{ if .nuon.install.inputs.root_domain }}{{ .nuon.install.inputs.root_domain }}{{ else }}sandbox-{{ .nuon.install.id | substr 0 15 }}.installs.example.com{{ end }}",
					"BucketName":  "vendor-{{ .nuon.install.id }}-service",
					"Inputs":      "{{ .nuon.inputs.inputs.cluster_version }}",
					"Environment": "production",
				},
			},
		},
	}
	require.NoError(t, cfg.parse())
}

func TestStackConfig_Parse_InvalidParameterValue(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr string
	}{
		{"unparseable template", "{{ if .nuon.install.inputs.foo }}no-end", "is not a valid template"},
		{"unknown function", "{{ .nuon.install.inputs.foo | nope }}", "is not a valid template"},
		{"empty value", "", "must not be empty"},
		{"sandbox output", "{{ .nuon.sandbox.outputs.cluster }}", "is not populated when the install stack is generated"},
		{"install stack output", "{{ .nuon.install_stack.outputs.region }}", "is not populated when the install stack is generated"},
		{"component output", "{{ .nuon.components.nlb.outputs.dns_name }}", "is not populated when the install stack is generated"},
		{"action output", "{{ .nuon.actions.setup.outputs.token }}", "is not populated when the install stack is generated"},
		{"legacy sandbox output", "{{ .nuon.install.sandbox.outputs.nuon_dns.public_domain.zone_id }}", "is not populated when the install stack is generated"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &StackConfig{
				CustomNestedStacks: []CustomNestedStack{
					{
						Name:        "my_stack",
						TemplateURL: "https://s3.amazonaws.com/bucket/template.yaml",
						Index:       0,
						Parameters: map[string]string{
							"Namespaces": tc.value,
						},
					},
				},
			}
			err := cfg.parse()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "parameter \"Namespaces\"")
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

func TestStackConfig_Parse_EmptyParameters(t *testing.T) {
	cfg := &StackConfig{
		Type:                    "aws-cloudformation",
		Name:                    "my-stack",
		Description:             "test stack",
		VPCNestedTemplateURL:    "https://s3.amazonaws.com/bucket/vpc.yaml",
		RunnerNestedTemplateURL: "https://s3.amazonaws.com/bucket/runner.yaml",
		CustomNestedStacks: []CustomNestedStack{
			{
				Name:        "my_stack",
				TemplateURL: "https://s3.amazonaws.com/bucket/template.yaml",
				Index:       0,
				Parameters:  map[string]string{},
			},
		},
	}
	require.NoError(t, cfg.parse())
}

func TestParseInstallInputReference(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantName  string
		wantError bool
	}{
		{"valid compact", "{{.nuon.install.inputs.namespaces}}", "namespaces", false},
		{"valid with spaces", "{{ .nuon.install.inputs.cluster_name }}", "cluster_name", false},
		{"valid with underscores", "{{.nuon.install.inputs.my_custom_param}}", "my_custom_param", false},
		{"missing dot", "{{nuon.install.inputs.namespaces}}", "", true},
		{"literal value", "some-value", "", true},
		{"wrong prefix", "{{.nuon.install.outputs.foo}}", "", true},
		{"empty", "", "", true},
		{"missing input name", "{{.nuon.install.inputs.}}", "", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			name, err := ParseInstallInputReference(tc.input)
			if tc.wantError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.wantName, name)
			}
		})
	}
}

func TestIsS3URL(t *testing.T) {
	tests := []struct {
		name  string
		url   string
		valid bool
	}{
		{"path-style global", "https://s3.amazonaws.com/bucket/key", true},
		{"path-style regional", "https://s3.us-west-2.amazonaws.com/bucket/key", true},
		{"virtual-hosted global", "https://bucket.s3.amazonaws.com/key", true},
		{"virtual-hosted regional", "https://bucket.s3.us-west-2.amazonaws.com/key", true},
		{"nuon-artifacts", "https://nuon-artifacts.s3.us-west-2.amazonaws.com/templates/k8s.yaml", true},
		{"non-s3 host", "https://example.com/template.yaml", false},
		{"http scheme", "http://s3.amazonaws.com/bucket/key", false},
		{"non-aws s3", "https://s3.example.com/bucket/key", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateTemplateURL(tc.url, "test_field")
			if tc.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestStackConfig_Parse_GCPCustomStacks(t *testing.T) {
	cfg := &StackConfig{
		Type:        "gcp-terraform",
		Name:        "my-stack",
		Description: "test stack",
		CustomNestedStacks: []CustomNestedStack{
			{Name: "preview_bucket", TemplateURL: "github.com/nuonco/install-stacks//gcp/modules/bucket", Index: 0},
		},
	}
	require.NoError(t, cfg.parse())
	require.Equal(t, "bucket", cfg.CustomNestedStacks[0].GCPModuleName())

	for _, forkURL := range []string{
		"github.com/acme/install-stacks//gcp/modules/bucket",
		"git::https://gitlab.com/acme/stacks.git//gcp/modules/bucket",
	} {
		cfg.CustomNestedStacks[0].TemplateURL = forkURL
		require.NoError(t, cfg.parse(), forkURL)
		require.Equal(t, "bucket", cfg.CustomNestedStacks[0].GCPModuleName(), forkURL)
	}

	for _, badURL := range []string{
		"https://example.com/stack.yaml",
		"github.com/nuonco/install-stacks//gcp/modules/",
		"github.com/nuonco/install-stacks//gcp/modules/bucket/extra",
	} {
		cfg.CustomNestedStacks[0].TemplateURL = badURL
		require.Error(t, cfg.parse(), badURL)
	}
}

func TestStackConfigParseRequiresGCPDNSName(t *testing.T) {
	cfg := &StackConfig{
		Type:        "gcp-terraform",
		Name:        "my-stack",
		Description: "test stack",
		CustomNestedStacks: []CustomNestedStack{
			{Name: "dns", TemplateURL: "github.com/nuonco/install-stacks//gcp/modules/dns", Index: 0},
		},
	}

	require.EqualError(t, cfg.parse(), "custom_nested_stacks[0] (dns): parameters.dns_name is required for the GCP dns module")
	cfg.CustomNestedStacks[0].Parameters = map[string]string{"dns_name": "app.example.com."}
	require.NoError(t, cfg.parse())
}

func TestStackConfig_ParseRejectsDuplicateCustomStackNames(t *testing.T) {
	cfg := &StackConfig{
		Type:        "gcp-terraform",
		Name:        "my-stack",
		Description: "test stack",
		CustomNestedStacks: []CustomNestedStack{
			{Name: "bucket", TemplateURL: "github.com/nuonco/install-stacks//gcp/modules/bucket", Index: 0},
			{Name: "bucket", TemplateURL: "github.com/nuonco/install-stacks//gcp/modules/bucket", Index: 1},
		},
	}

	err := cfg.parse()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "each stack must have a unique name")
}
