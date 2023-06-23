package pipeline

import (
	"context"
	"reflect"

	"github.com/go-faker/faker/v4"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
)

func fakePipelineExecFn(v reflect.Value) (interface{}, error) {
	return func(ctx context.Context, l hclog.Logger, ui terminal.UI) ([]byte, error) {
		return []byte("hello world"), nil
	}, nil
}

func fakePipelineCallbackFn(v reflect.Value) (interface{}, error) {
	return func(ctx context.Context, l hclog.Logger, ui terminal.UI, byt []byte) error {
		return nil
	}, nil
}

func init() {
	_ = faker.AddProvider("pipelineCallbackFn", fakePipelineCallbackFn)
	_ = faker.AddProvider("pipelineExecFn", fakePipelineExecFn)
}
