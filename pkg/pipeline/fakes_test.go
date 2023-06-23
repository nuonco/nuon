package pipeline

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	"github.com/stretchr/testify/assert"
)

func Test_fakePipelineExecFn(t *testing.T) {
	t.Run("ensure pipeline exec func is returned", func(t *testing.T) {
		val, err := fakePipelineExecFn(reflect.ValueOf("anything"))
		assert.NoError(t, err)

		fn, ok := val.(func(context.Context, hclog.Logger, terminal.UI) ([]byte, error))
		assert.True(t, ok)

		res, err := fn(nil, nil, nil)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})
}

func Test_fakePipelineCallbackFn(t *testing.T) {
	t.Run("ensure pipeline callback func is returned", func(t *testing.T) {
		val, err := fakePipelineCallbackFn(reflect.ValueOf("anything"))
		assert.NoError(t, err)

		fn, ok := val.(func(context.Context, hclog.Logger, terminal.UI, []byte) error)
		assert.True(t, ok)

		err = fn(nil, nil, nil, nil)
		assert.NoError(t, err)
	})
}
