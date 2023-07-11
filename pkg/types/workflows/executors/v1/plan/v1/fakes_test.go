package planv1

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_fakeSandboxInputAccountSettings(t *testing.T) {
	input, err := fakeSandboxInputAccountSettings(reflect.ValueOf("anything"))
	assert.NoError(t, err)
	assert.NotNil(t, input)

	sandboxInput, ok := input.(*SandboxInput_Aws)
	assert.True(t, ok)
	assert.NotEmpty(t, sandboxInput.Aws.AccountId)
	assert.NotEmpty(t, sandboxInput.Aws.RoleArn)
	assert.NotEmpty(t, sandboxInput.Aws.Region)
}
