package examples

import (
	"context"
)

type QueueActivity struct{}

// queueAction
// @temporal-gen-v2 activity
// @as-wrapper
// @by-field Method
func (a *QueueActivity) queueAction(ctx context.Context, method string) (string, error) {
	return "done", nil
}
