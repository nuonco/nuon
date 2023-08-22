package executev1

import (
	reflect "reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFakeTerraformOutputs(t *testing.T) {
	obj, err := fakeTerraformOutputs(reflect.ValueOf("anything"))
	require.NoError(t, err)
	require.NotNil(t, obj)
}
