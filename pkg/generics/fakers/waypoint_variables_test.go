package fakers

import (
	"reflect"
	"testing"

	planv1 "github.com/powertoolsdev/mono/pkg/types/workflows/executors/v1/plan/v1"
	"github.com/stretchr/testify/assert"
)

func Test_fakeWaypointVariables(t *testing.T) {
	resp, err := fakeWaypointVariables(reflect.ValueOf("anything"))
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	waypointVariables, ok := resp.([]*planv1.WaypointVariable)
	assert.True(t, ok)
	for _, envVar := range waypointVariables {
		err = envVar.Validate()
		assert.NoError(t, err)
	}

}
