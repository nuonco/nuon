// Package badtag is a fixture for TestGenerateUnknownTagFails: the tag
// referenced below is not declared in this directory's temporal-gen.yaml, so
// generating here must fail rather than silently apply no defaults.
package badtag

import "context"

// @temporal-gen-v2 activity
// @tag does-not-exist
func BadTagActivity(ctx context.Context, input string) (string, error) {
	return input, nil
}
