package planv1

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_fakePlanConfigs(t *testing.T) {
	cfg, err := fakePlanConfigs(reflect.ValueOf("anything"))
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	cfgs, ok := cfg.([]*Config)
	assert.True(t, ok)

	for _, cfg := range cfgs {
		assert.NoError(t, cfg.Validate())
	}
}

func Test_fakeEnvVars(t *testing.T) {
	resp, err := fakeEnvVars(reflect.ValueOf("anything"))
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	envVars, ok := resp.([]*EnvVar)
	assert.True(t, ok)
	for _, envVar := range envVars {
		err = envVar.Validate()
		assert.NoError(t, err)
	}

}

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

func Test_fakeWaypointVariables(t *testing.T) {
	resp, err := fakeWaypointVariables(reflect.ValueOf("anything"))
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	waypointVariables, ok := resp.([]*WaypointVariable)
	assert.True(t, ok)
	for _, envVar := range waypointVariables {
		err = envVar.Validate()
		assert.NoError(t, err)
	}

}
