package parser

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/config"
	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/tags"
)

func testConfig(t *testing.T) *tags.Config {
	t.Helper()
	path := filepath.Join(t.TempDir(), "temporal-gen.yaml")
	require.NoError(t, os.WriteFile(path, []byte(`
version: 1
defaults:
  start-to-close-timeout: 1m
tags:
  db-read:
    start-to-close-timeout: 30s
    max-retries: 3
  bulk:
    start-to-close-timeout: 1h
    heartbeat-timeout: 1m
  critical:
    max-retries: 870
`), 0o644))

	cfg, err := tags.Load(path)
	require.NoError(t, err)
	return cfg
}

func marker(kind string) string { return "// @" + config.AnnotationPrefix + " " + kind }

func TestParseWithTags(t *testing.T) {
	cfg := testConfig(t)

	t.Run("defaults apply with no tags", func(t *testing.T) {
		a, err := ParseWithTags([]string{marker("activity")}, cfg)
		require.NoError(t, err)
		assert.Equal(t, time.Minute, a.ActivityOpts.StartToCloseTimeout)
		assert.Empty(t, a.Tags)
	})

	t.Run("tag overrides defaults", func(t *testing.T) {
		a, err := ParseWithTags([]string{marker("activity"), "// @tag db-read"}, cfg)
		require.NoError(t, err)
		assert.Equal(t, 30*time.Second, a.ActivityOpts.StartToCloseTimeout)
		assert.Equal(t, 3, a.ActivityOpts.MaxRetries)
		assert.True(t, a.ActivityOpts.RetryPolicy)
		assert.Equal(t, []string{"db-read"}, a.Tags)
	})

	t.Run("distinct tags both contribute", func(t *testing.T) {
		a, err := ParseWithTags([]string{
			marker("activity"),
			"// @tag bulk",
			"// @tag critical",
		}, cfg)
		require.NoError(t, err)
		assert.Equal(t, time.Hour, a.ActivityOpts.StartToCloseTimeout)
		assert.Equal(t, time.Minute, a.ActivityOpts.HeartbeatTimeout)
		assert.Equal(t, 870, a.ActivityOpts.MaxRetries)
	})

	t.Run("explicit annotation beats tag", func(t *testing.T) {
		a, err := ParseWithTags([]string{
			marker("activity"),
			"// @tag db-read",
			"// @start-to-close-timeout 10s",
		}, cfg)
		require.NoError(t, err)
		assert.Equal(t, 10*time.Second, a.ActivityOpts.StartToCloseTimeout)
		assert.Equal(t, 3, a.ActivityOpts.MaxRetries)
	})

	t.Run("explicit annotation order does not matter", func(t *testing.T) {
		// @tag appearing after the annotation it would override must still lose.
		a, err := ParseWithTags([]string{
			marker("activity"),
			"// @start-to-close-timeout 10s",
			"// @tag db-read",
		}, cfg)
		require.NoError(t, err)
		assert.Equal(t, 10*time.Second, a.ActivityOpts.StartToCloseTimeout)
	})

	t.Run("later tag wins when two tags touch the same attribute", func(t *testing.T) {
		a, err := ParseWithTags([]string{
			marker("activity"),
			"// @tag db-read",
			"// @tag critical",
		}, cfg)
		require.NoError(t, err)
		assert.Equal(t, 870, a.ActivityOpts.MaxRetries)
	})

	t.Run("in-code config behaves the same as a file", func(t *testing.T) {
		inCode := &tags.Config{
			Tags: map[string]*tags.Attrs{
				"db-read": {StartToCloseTimeout: "30s"},
			},
		}
		require.NoError(t, inCode.Validate())

		a, err := ParseWithTags([]string{marker("activity"), "// @tag db-read"}, inCode)
		require.NoError(t, err)
		assert.Equal(t, 30*time.Second, a.ActivityOpts.StartToCloseTimeout)
	})

	t.Run("tag on a workflow errors", func(t *testing.T) {
		_, err := ParseWithTags([]string{marker("workflow"), "// @tag critical"}, cfg)
		requireTagError(t, err)
		assert.Contains(t, err.Error(), "@tag is only supported on activities, found on workflow")
	})

	t.Run("workflow without tags is unaffected", func(t *testing.T) {
		a, err := ParseWithTags([]string{marker("workflow")}, cfg)
		require.NoError(t, err)
		// the activity-only defaults block must not leak onto workflows
		assert.Zero(t, a.WorkflowOpts.ExecutionTimeout)
	})

	t.Run("unknown tag errors", func(t *testing.T) {
		_, err := ParseWithTags([]string{marker("activity"), "// @tag nope"}, cfg)
		requireTagError(t, err)
		assert.Contains(t, err.Error(), `unknown tag "nope"`)
	})

	t.Run("duplicate tag errors", func(t *testing.T) {
		_, err := ParseWithTags([]string{
			marker("activity"),
			"// @tag db-read",
			"// @tag db-read",
		}, cfg)
		requireTagError(t, err)
		assert.Contains(t, err.Error(), `tag "db-read" set twice`)
	})

	t.Run("tags without a config errors", func(t *testing.T) {
		_, err := ParseWithTags([]string{marker("activity"), "// @tag db-read"}, nil)
		requireTagError(t, err)
		assert.Contains(t, err.Error(), "no tag config was found")
	})

	t.Run("missing name errors", func(t *testing.T) {
		_, err := ParseWithTags([]string{marker("activity"), "// @tag"}, cfg)
		requireTagError(t, err)
		assert.Contains(t, err.Error(), "missing name for @tag")
	})

	t.Run("unannotated comments stay nil", func(t *testing.T) {
		a, err := ParseWithTags([]string{"// just a normal comment"}, cfg)
		require.NoError(t, err)
		assert.Nil(t, a)
	})

	t.Run("validation still runs after tag merge", func(t *testing.T) {
		_, err := ParseWithTags([]string{marker("activity"), "// @by-field-only"}, cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "@by-field-only requires @by-field")
	})
}

// requireTagError asserts the failure is a *TagError, which is what makes it
// fatal in non-strict generation runs.
func requireTagError(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	var tagErr *TagError
	require.True(t, errors.As(err, &tagErr), "expected a *TagError, got %T: %v", err, err)
}
