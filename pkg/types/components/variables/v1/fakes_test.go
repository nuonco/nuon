package variablesv1

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_fakeVariables(t *testing.T) {
	cfg, err := fakeVariables(reflect.ValueOf("anything"))
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	cfgs, ok := cfg.([]*Variable)
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

func Test_fakeHelmValues(t *testing.T) {
	resp, err := fakeHelmValues(reflect.ValueOf("anything"))
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	helmValues, ok := resp.([]*HelmValue)
	assert.True(t, ok)
	for _, envVar := range helmValues {
		err = envVar.Validate()
		assert.NoError(t, err)
	}
}

func Test_fakeTerraformVariables(t *testing.T) {
	resp, err := fakeTerraformVariables(reflect.ValueOf("anything"))
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	tfVariables, ok := resp.([]*TerraformVariable)
	assert.True(t, ok)
	for _, envVar := range tfVariables {
		err = envVar.Validate()
		assert.NoError(t, err)
	}
}
