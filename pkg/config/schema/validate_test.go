package schema

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
)

func TestValidate_StackWithoutAdditionalNestedStacks(t *testing.T) {
	cfg := &config.AppConfig{
		Stack: &config.StackConfig{
			Type:                    "aws-cloudformation",
			Name:                    "test-stack",
			Description:             "test",
			VPCNestedTemplateURL:    "https://example.com/vpc.yaml",
			RunnerNestedTemplateURL: "https://example.com/runner.yaml",
		},
	}

	errs, err := Validate(context.Background(), cfg)
	require.NoError(t, err)

	for _, e := range errs {
		assert.NotContains(t, e.String(), "additional_nested_stacks",
			"additional_nested_stacks should not cause validation errors when omitted, got: %s", e.String())
	}
}

func TestValidate_StackWithAdditionalNestedStacks(t *testing.T) {
	cfg := &config.AppConfig{
		Stack: &config.StackConfig{
			Type:                    "aws-cloudformation",
			Name:                    "test-stack",
			Description:             "test",
			VPCNestedTemplateURL:    "https://example.com/vpc.yaml",
			RunnerNestedTemplateURL: "https://example.com/runner.yaml",
			AdditionalNestedStacks: []config.AdditionalNestedStack{
				{Name: "k8s_namespaces", TemplateURL: "https://example.com/ns.yaml", Index: 0},
			},
		},
	}

	errs, err := Validate(context.Background(), cfg)
	require.NoError(t, err)

	for _, e := range errs {
		assert.NotContains(t, e.String(), "additional_nested_stacks",
			"additional_nested_stacks should not cause validation errors when valid, got: %s", e.String())
	}
}
