// Package libtags is a fixture for the lib.Generate tests. Its temporal-gen.yaml
// declares db-read with deliberately different values than the in-code configs
// those tests pass, so the generated output shows which config was actually
// used.
package libtags

import "context"

// @temporal-gen-v2 activity
// @tag db-read
func TaggedActivity(ctx context.Context, input string) (string, error) {
	return input, nil
}
