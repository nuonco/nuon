package examples

// Demonstrates @label key value, which pulls default options from
// temporal-gen.yaml in this directory. Labels apply to activities only.
// See AGENTS.md for the precedence rules.

import "context"

// UnlabeledActivity carries no labels, so it still picks up the `defaults`
// block (start-to-close-timeout 1m) from temporal-gen.yaml.
// @temporal-gen-v2 activity
func UnlabeledActivity(ctx context.Context, input string) (string, error) {
	return "result", nil
}

// SingleLabelActivity inherits start-to-close-timeout 30s and max-retries 3
// from access=db-only, overriding the 1m default.
// @temporal-gen-v2 activity
// @label access db-only
func SingleLabelActivity(ctx context.Context, id string) (string, error) {
	return "user-" + id, nil
}

// MultiLabelActivity sets two different label keys. access=bulk supplies the
// timeouts and heartbeat; tier=critical supplies max-retries 870. Distinct
// keys set disjoint attributes here, so both contribute in full.
// @temporal-gen-v2 activity
// @label access bulk
// @label tier critical
func MultiLabelActivity(ctx context.Context, input int) (int, error) {
	return input * 2, nil
}

// ConflictingLabelActivity shows ordering when two keys touch the same
// attribute: access=db-only sets max-retries 3 and tier=critical sets 870.
// tier is listed second, so 870 wins.
// @temporal-gen-v2 activity
// @label access db-only
// @label tier critical
func ConflictingLabelActivity(ctx context.Context, input string) error {
	return nil
}

// LabelOverriddenActivity shows an explicit annotation beating its label:
// access=db-only asks for 30s, but @start-to-close-timeout 10s is applied
// after the label defaults and therefore wins. max-retries 3 still comes from
// the label.
// @temporal-gen-v2 activity
// @label access db-only
// @start-to-close-timeout 10s
func LabelOverriddenActivity(ctx context.Context, input string) error {
	return nil
}
