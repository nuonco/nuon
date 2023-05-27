package fakers

import (
	"context"
	"reflect"

	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	"go.uber.org/zap"
)

func fakePipelineExecFn(v reflect.Value) (interface{}, error) {
	return func(ctx context.Context, l *zap.Logger, ui terminal.UI) ([]byte, error) {
		return []byte("hello world"), nil
	}, nil
}

func fakePipelineCallbackFn(v reflect.Value) (interface{}, error) {
	return func(ctx context.Context, l *zap.Logger, ui terminal.UI, byt []byte) error {
		return nil
	}, nil
}
