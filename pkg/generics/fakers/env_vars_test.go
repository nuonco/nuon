package fakers

import (
	"reflect"
	"testing"

	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	"github.com/stretchr/testify/assert"
)

func Test_fakeEnvVars(t *testing.T) {
	resp, err := fakeEnvVars(reflect.ValueOf("anything"))
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	envVars, ok := resp.([]*planv1.EnvVar)
	assert.True(t, ok)
	for _, envVar := range envVars {
		err = envVar.Validate()
		assert.NoError(t, err)
	}

}
