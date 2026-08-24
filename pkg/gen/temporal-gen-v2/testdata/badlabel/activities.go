// Package badlabel is a fixture for TestGenerateUnknownLabelFails: the label
// key referenced below is not declared in this directory's temporal-gen.yaml,
// so generating here must fail rather than silently apply no defaults.
package badlabel

import "context"

// @temporal-gen-v2 activity
// @label does-not-exist yes
func BadLabelActivity(ctx context.Context, input string) (string, error) {
	return input, nil
}
