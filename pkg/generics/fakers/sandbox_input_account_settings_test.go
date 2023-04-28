package fakers

import (
	"reflect"
	"testing"

	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	"github.com/stretchr/testify/assert"
)

func Test_fakeSandboxInputAccountSettings(t *testing.T) {
	input, err := fakeSandboxInputAccountSettings(reflect.ValueOf("anything"))
	assert.NoError(t, err)
	assert.NotNil(t, input)

	sandboxInput, ok := input.(*planv1.SandboxInput_Aws)
	assert.True(t, ok)
	assert.NotEmpty(t, sandboxInput.Aws.AccountId)
	assert.NotEmpty(t, sandboxInput.Aws.RoleArn)
	assert.NotEmpty(t, sandboxInput.Aws.Region)
}
