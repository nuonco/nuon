package parser

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/config"
	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/internal/labels"
)

func testConfig(t *testing.T) *labels.Config {
	t.Helper()
	path := filepath.Join(t.TempDir(), "temporal-gen.yaml")
	require.NoError(t, os.WriteFile(path, []byte(`
version: 1
defaults:
  start-to-close-timeout: 1m
labels:
  access:
    values:
      db-only:
        start-to-close-timeout: 30s
        max-retries: 3
      bulk:
        start-to-close-timeout: 1h
        heartbeat-timeout: 1m
  tier:
    values:
      critical:
        max-retries: 870
`), 0o644))

	cfg, err := labels.Load(path)
	require.NoError(t, err)
	return cfg
}

func marker(kind string) string { return "// @" + config.AnnotationPrefix + " " + kind }

func TestParseWithLabels(t *testing.T) {
	cfg := testConfig(t)

	t.Run("defaults apply with no labels", func(t *testing.T) {
		a, err := ParseWithLabels([]string{marker("activity")}, cfg)
		require.NoError(t, err)
		assert.Equal(t, time.Minute, a.ActivityOpts.StartToCloseTimeout)
		assert.Empty(t, a.Labels)
	})

	t.Run("label overrides defaults", func(t *testing.T) {
		a, err := ParseWithLabels([]string{marker("activity"), "// @label access db-only"}, cfg)
		require.NoError(t, err)
		assert.Equal(t, 30*time.Second, a.ActivityOpts.StartToCloseTimeout)
		assert.Equal(t, 3, a.ActivityOpts.MaxRetries)
		assert.True(t, a.ActivityOpts.RetryPolicy)
		assert.Equal(t, []labels.Pair{{Key: "access", Value: "db-only"}}, a.Labels)
	})

	t.Run("distinct keys both contribute", func(t *testing.T) {
		a, err := ParseWithLabels([]string{
			marker("activity"),
			"// @label access bulk",
			"// @label tier critical",
		}, cfg)
		require.NoError(t, err)
		assert.Equal(t, time.Hour, a.ActivityOpts.StartToCloseTimeout)
		assert.Equal(t, time.Minute, a.ActivityOpts.HeartbeatTimeout)
		assert.Equal(t, 870, a.ActivityOpts.MaxRetries)
	})

	t.Run("explicit annotation beats label", func(t *testing.T) {
		a, err := ParseWithLabels([]string{
			marker("activity"),
			"// @label access db-only",
			"// @start-to-close-timeout 10s",
		}, cfg)
		require.NoError(t, err)
		assert.Equal(t, 10*time.Second, a.ActivityOpts.StartToCloseTimeout)
		assert.Equal(t, 3, a.ActivityOpts.MaxRetries)
	})

	t.Run("explicit annotation order does not matter", func(t *testing.T) {
		// @label appearing after the annotation it would override must still lose.
		a, err := ParseWithLabels([]string{
			marker("activity"),
			"// @start-to-close-timeout 10s",
			"// @label access db-only",
		}, cfg)
		require.NoError(t, err)
		assert.Equal(t, 10*time.Second, a.ActivityOpts.StartToCloseTimeout)
	})

	t.Run("later key wins when two keys touch the same attribute", func(t *testing.T) {
		a, err := ParseWithLabels([]string{
			marker("activity"),
			"// @label access db-only",
			"// @label tier critical",
		}, cfg)
		require.NoError(t, err)
		assert.Equal(t, 870, a.ActivityOpts.MaxRetries)
	})

	t.Run("label on a workflow errors", func(t *testing.T) {
		_, err := ParseWithLabels([]string{marker("workflow"), "// @label tier critical"}, cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "@label is only supported on activities, found on workflow")
	})

	t.Run("workflow without labels is unaffected", func(t *testing.T) {
		a, err := ParseWithLabels([]string{marker("workflow")}, cfg)
		require.NoError(t, err)
		// the activity-only defaults block must not leak onto workflows
		assert.Zero(t, a.WorkflowOpts.ExecutionTimeout)
	})

	t.Run("unknown key errors", func(t *testing.T) {
		_, err := ParseWithLabels([]string{marker("activity"), "// @label nope x"}, cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `unknown label key "nope"`)
	})

	t.Run("unknown value errors", func(t *testing.T) {
		_, err := ParseWithLabels([]string{marker("activity"), "// @label access sideways"}, cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `unknown value "sideways" for label key "access"`)
	})

	t.Run("duplicate key errors", func(t *testing.T) {
		_, err := ParseWithLabels([]string{
			marker("activity"),
			"// @label access db-only",
			"// @label access bulk",
		}, cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `label key "access" set twice`)
	})

	t.Run("labels without a config errors", func(t *testing.T) {
		_, err := ParseWithLabels([]string{marker("activity"), "// @label access db-only"}, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no temporal-gen.yaml was found")
	})

	t.Run("missing value errors", func(t *testing.T) {
		_, err := ParseWithLabels([]string{marker("activity"), "// @label access"}, cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "missing key and value for @label")
	})

	t.Run("unannotated comments stay nil", func(t *testing.T) {
		a, err := ParseWithLabels([]string{"// just a normal comment"}, cfg)
		require.NoError(t, err)
		assert.Nil(t, a)
	})

	t.Run("validation still runs after label merge", func(t *testing.T) {
		_, err := ParseWithLabels([]string{marker("activity"), "// @by-field-only"}, cfg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "@by-field-only requires @by-field")
	})
}
