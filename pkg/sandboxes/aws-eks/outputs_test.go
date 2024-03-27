package awseks

import (
	"reflect"
	"testing"

	"github.com/powertoolsdev/mono/pkg/sandboxes"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestParseTerraformOutputs(t *testing.T) {
	fakeData, err := fakeOutputs(reflect.ValueOf("anything"))
	assert.Nil(t, err)
	outputs, ok := fakeData.(*structpb.Struct)
	assert.True(t, ok)

	parsed, err := ParseTerraformOutputs(outputs)
	assert.Nil(t, err)
	assert.NotNil(t, parsed)
	assert.NoError(t, parsed.Validate())
}

func TestToStringSlice(t *testing.T) {
	strVals := sandboxes.ToStringSlice([]interface{}{"abc"})
	assert.Equal(t, "abc", strVals[0])
}
