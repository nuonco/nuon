package examples

// Demonstrates @label key value, which pulls default options from
// temporal-gen.yaml in this directory. See AGENTS.md for the precedence rules.

import (
	"context"

	"go.temporal.io/sdk/workflow"
)

// UnlabeledActivity carries no labels, so it still picks up the `defaults`
// block (start-to-close-timeout 1m) from temporal-gen.yaml.
// @temporal-gen-v2 activity
func UnlabeledActivity(ctx context.Context, input string) (string, error) {
	return "result", nil
}

// SingleLabelActivity inherits start-to-close-timeout 30s and max-retries 3
// from access=read-only, overriding the 1m default.
// @temporal-gen-v2 activity
// @label access read-only
func SingleLabelActivity(ctx context.Context, id string) (string, error) {
	return "user-" + id, nil
}

// MultiLabelActivity sets two different label keys. access=bulk supplies the
// timeouts and heartbeat; tier=critical supplies max-retries 870. Distinct
// keys never conflict, so both contribute in full.
// @temporal-gen-v2 activity
// @label access bulk
// @label tier critical
func MultiLabelActivity(ctx context.Context, input int) (int, error) {
	return input * 2, nil
}

// ConflictingLabelActivity shows ordering between keys that do collide: both
// access=read-only and access=bulk would set start-to-close-timeout, but a key
// may only be set once, so this instead pairs read-only with tier=best-effort.
// tier=best-effort's max-retries 3 is applied after access=read-only's, and
// they agree; had they differed, the later line would win.
// @temporal-gen-v2 activity
// @label access read-only
// @label tier best-effort
func ConflictingLabelActivity(ctx context.Context, input string) error {
	return nil
}

// LabelOverriddenActivity shows an explicit annotation beating its label:
// access=read-only asks for 30s, but @start-to-close-timeout 10s is applied
// after the label defaults and therefore wins. max-retries 3 still comes from
// the label.
// @temporal-gen-v2 activity
// @label access read-only
// @start-to-close-timeout 10s
func LabelOverriddenActivity(ctx context.Context, input string) error {
	return nil
}

// LabeledWorkflow uses tier=critical from the workflow side, so it picks up
// the workflow block instead of the activity one: execution-timeout 24h and a
// tier memo.
// @temporal-gen-v2 workflow
// @label tier critical
func LabeledWorkflow(ctx workflow.Context, input string) (string, error) {
	return "done", nil
}
