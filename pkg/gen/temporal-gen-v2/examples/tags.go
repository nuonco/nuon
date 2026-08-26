package examples

// Demonstrates @tag name, which pulls default options from temporal-gen.yaml
// in this directory. Tags apply to activities only.
// See AGENTS.md for the precedence rules.

import "context"

// UntaggedActivity carries no tags, so it still picks up the `defaults` block
// (start-to-close-timeout 1m) from temporal-gen.yaml.
// @temporal-gen-v2 activity
func UntaggedActivity(ctx context.Context, input string) (string, error) {
	return "result", nil
}

// SingleTagActivity inherits start-to-close-timeout 30s and max-retries 3 from
// db-read, overriding the 1m default.
// @temporal-gen-v2 activity
// @tag db-read
func SingleTagActivity(ctx context.Context, id string) (string, error) {
	return "user-" + id, nil
}

// MultiTagActivity sets two tags. bulk supplies the timeouts and heartbeat;
// critical supplies max-retries 870. They set disjoint attributes here, so both
// contribute in full.
// @temporal-gen-v2 activity
// @tag bulk
// @tag critical
func MultiTagActivity(ctx context.Context, input int) (int, error) {
	return input * 2, nil
}

// ConflictingTagActivity shows ordering when two tags touch the same
// attribute: db-read sets max-retries 3 and critical sets 870. critical is
// listed second, so 870 wins.
// @temporal-gen-v2 activity
// @tag db-read
// @tag critical
func ConflictingTagActivity(ctx context.Context, input string) error {
	return nil
}

// TagOverriddenActivity shows an explicit annotation beating its tag: db-read
// asks for 30s, but @start-to-close-timeout 10s is applied after the tag
// defaults and therefore wins. max-retries 3 still comes from the tag.
// @temporal-gen-v2 activity
// @tag db-read
// @start-to-close-timeout 10s
func TagOverriddenActivity(ctx context.Context, input string) error {
	return nil
}
