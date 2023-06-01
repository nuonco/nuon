package fakers

import (
	"context"
	"reflect"

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
