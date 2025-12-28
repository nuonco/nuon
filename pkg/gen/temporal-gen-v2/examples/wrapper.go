package examples

import (
	"context"
)

type WrapperActivity struct{}

// wrapperActivityAction demonstrates the wrapper generation
// @temporal-gen-v2 activity
// @as-wrapper
// @by-field id
func (a *WrapperActivity) wrapperActivityAction(ctx context.Context, id string, count int) (string, error) {
	return "result", nil
}
