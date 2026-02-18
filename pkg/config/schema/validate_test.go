package schema

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
)

func TestValidate_StackWithoutCustomNestedStacks(t *testing.T) {
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
		assert.NotContains(t, e.String(), "custom_nested_stacks",
			"custom_nested_stacks should not cause validation errors when omitted, got: %s", e.String())
	}
}
