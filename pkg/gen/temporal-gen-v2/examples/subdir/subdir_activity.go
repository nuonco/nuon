package subdir

import (
	"context"
)

// SubdirActivity
// @temporal-gen-v2 activity
func SubdirActivity(ctx context.Context, input string) (string, error) {
	return "subdir", nil
}
