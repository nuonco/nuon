package pkgc

import "context"

// PkgcActivity is a base activity with no dependencies.
// @temporal-gen-v2 activity
func PkgcActivity(ctx context.Context, input string) (string, error) {
	return input, nil
}
