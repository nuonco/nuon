package pipeline

import (
	"context"
	"reflect"

	"github.com/go-faker/faker/v4"
	"github.com/hashicorp/go-hclog"
)

func fakePipelineExecFn(v reflect.Value) (interface{}, error) {
	return func(ctx context.Context, l hclog.Logger) ([]byte, error) {
		return []byte("hello world"), nil
	}, nil
}

func fakePipelineCallbackFn(v reflect.Value) (interface{}, error) {
	return func(ctx context.Context, l hclog.Logger, byt []byte) error {
		return nil
	}, nil
}

func init() {
	_ = faker.AddProvider("pipelineCallbackFn", fakePipelineCallbackFn)
	_ = faker.AddProvider("pipelineExecFn", fakePipelineExecFn)
}
