package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStackConfig_Parse_NoCustomNestedStacks(t *testing.T) {
	cfg := &StackConfig{
		Type:        "aws-cloudformation",
		Name:        "my-stack",
		Description: "test stack",
	}
	require.NoError(t, cfg.parse())
}

func TestStackConfig_Parse_EmptyCustomNestedStacks(t *testing.T) {
	cfg := &StackConfig{
		Type:               "aws-cloudformation",
		Name:               "my-stack",
		Description:        "test stack",
		CustomNestedStacks: []CustomNestedStack{},
	}
	require.NoError(t, cfg.parse())
}

func TestStackConfig_Parse_ValidCustomNestedStacks(t *testing.T) {
	cfg := &StackConfig{
		Type:        "aws-cloudformation",
		Name:        "my-stack",
		Description: "test stack",
		CustomNestedStacks: []CustomNestedStack{
			{Name: "k8s_namespaces", TemplateURL: "https://s3.amazonaws.com/bucket/template.yaml", Index: 0},
		},
	}
	require.NoError(t, cfg.parse())
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

func TestStackConfig_Parse_InvalidTemplateURL(t *testing.T) {
	cfg := &StackConfig{
		CustomNestedStacks: []CustomNestedStack{
			{Name: "my_stack", TemplateURL: "not-a-url", Index: 0},
		},
	}
	err := cfg.parse()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing scheme or host")
}

func TestStackConfig_Parse_TemplateURLMissingScheme(t *testing.T) {
	cfg := &StackConfig{
		CustomNestedStacks: []CustomNestedStack{
			{Name: "my_stack", TemplateURL: "example.com/template.yaml", Index: 0},
		},
	}
	err := cfg.parse()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing scheme or host")
}

func TestStackConfig_Parse_NonS3URL(t *testing.T) {
	cfg := &StackConfig{
		CustomNestedStacks: []CustomNestedStack{
			{Name: "my_stack", TemplateURL: "https://example.com/template.yaml", Index: 0},
		},
	}
	err := cfg.parse()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not an S3 URL")
}

func TestStackConfig_Parse_HttpS3URLRejected(t *testing.T) {
	cfg := &StackConfig{
		CustomNestedStacks: []CustomNestedStack{
			{Name: "my_stack", TemplateURL: "http://s3.amazonaws.com/bucket/template.yaml", Index: 0},
		},
	}
	err := cfg.parse()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not an S3 URL")
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
		Type:        "aws-cloudformation",
		Name:        "my-stack",
		Description: "test stack",
		CustomNestedStacks: []CustomNestedStack{
			{
				Name:        "k8s_namespaces",
				TemplateURL: "https://s3.amazonaws.com/bucket/template.yaml",
				Index:       0,
				Parameters: map[string]string{
					"Namespaces":  "{{.nuon.install.inputs.namespaces}}",
					"ClusterName": "{{ .nuon.install.inputs.cluster_name }}",
				},
			},
		},
	}
	require.NoError(t, cfg.parse())
}

func TestStackConfig_Parse_InvalidParameterValue(t *testing.T) {
	cfg := &StackConfig{
		CustomNestedStacks: []CustomNestedStack{
			{
				Name:        "my_stack",
				TemplateURL: "https://s3.amazonaws.com/bucket/template.yaml",
				Index:       0,
				Parameters: map[string]string{
					"Namespaces": "some-literal-value",
				},
			},
		},
	}
	err := cfg.parse()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parameter \"Namespaces\"")
	assert.Contains(t, err.Error(), "must be a template reference")
}

func TestStackConfig_Parse_EmptyParameters(t *testing.T) {
	cfg := &StackConfig{
		Type:        "aws-cloudformation",
		Name:        "my-stack",
		Description: "test stack",
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
