package awseks

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

func Test_fakeOutputs(t *testing.T) {
	outputs, err := fakeOutputs(reflect.ValueOf("anything"))
	assert.NoError(t, err)
	assert.NotNil(t, outputs)

	// ensure that the outputs can be decoded
	parsedOutputs, err := ParseTerraformOutputs(outputs.(*structpb.Struct))
	assert.NoError(t, err)
	assert.NoError(t, parsedOutputs.Validate())
}

func Test_fakeStringSliceAsInt(t *testing.T) {
	vals, err := fakeStringSliceAsInt(reflect.ValueOf("anything"))
	assert.NoError(t, err)
	assert.NotNil(t, vals)

	// ensure that the outputs can be decoded
	intVals, ok := vals.([]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, intVals)
}

func Test_fakeDomain(t *testing.T) {
	domain, err := fakeDomain(reflect.ValueOf("anything"))
	assert.NoError(t, err)
	assert.NotEmpty(t, domain)

	domainStr, ok := domain.(string)
	assert.True(t, ok)
	assert.NotEmpty(t, domainStr)
}
