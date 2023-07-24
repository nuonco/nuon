package installsv1

import (
	reflect "reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

func Test_fakeTerraformOutputs(t *testing.T) {
	resp, err := fakeTerraformOutputs(reflect.ValueOf("anything"))
	assert.NoError(t, err)
	assert.NotNil(t, resp)

	obj, ok := resp.(*structpb.Struct)
	assert.True(t, ok)

	vals := obj.AsMap()
	assert.Equal(t, "value", vals["key"])
	assert.Equal(t, "value", vals["obj"].(map[string]interface{})["key"])
}
