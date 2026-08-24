package labels

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeConfig(t *testing.T, dir, body string) string {
	t.Helper()
	path := filepath.Join(dir, "temporal-gen.yaml")
	require.NoError(t, os.WriteFile(path, []byte(body), 0o644))
	return path
}

const validConfig = `
version: 1
defaults:
  activity:
    start-to-close-timeout: 1m
labels:
  access:
    description: how the activity reaches its data
    values:
      read-only:
        activity:
          start-to-close-timeout: 30s
          max-retries: 3
      bulk:
        activity:
          start-to-close-timeout: 1h
          heartbeat-timeout: 1m
  tier:
    values:
      critical:
        activity:
          max-retries: 870
        workflow:
          execution-timeout: 24h
          memo:
            tier: critical
            owner: platform
`

func TestLoadValid(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, validConfig)

	cfg, err := Load(path)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, path, cfg.Path())
	assert.Equal(t, []string{"access", "tier"}, cfg.Keys())
	assert.Equal(t, []string{"bulk", "read-only"}, cfg.Values("access"))
}

func TestLoadRejectsUnknownAttribute(t *testing.T) {
	dir := t.TempDir()
	// "start-to-close" is a typo for "start-to-close-timeout" and must not be
	// silently ignored.
	path := writeConfig(t, dir, `
version: 1
labels:
  access:
    values:
      oops:
        activity:
          start-to-close: 30s
`)

	_, err := Load(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "start-to-close")
}

func TestLoadRejectsStructuralAttribute(t *testing.T) {
	dir := t.TempDir()
	// @as-wrapper is deliberately not settable from a label.
	path := writeConfig(t, dir, `
version: 1
labels:
  access:
    values:
      oops:
        activity:
          as-wrapper: true
`)

	_, err := Load(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "as-wrapper")
}

func TestLoadRejectsBadDuration(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, `
version: 1
labels:
  access:
    values:
      oops:
        activity:
          start-to-close-timeout: 30 seconds
`)

	_, err := Load(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid duration")
}

const minimalLabel = "labels:\n  tier:\n    values:\n      critical:\n        activity:\n          max-retries: 1\n"

func TestValidateVersion(t *testing.T) {
	dir := t.TempDir()

	_, err := Load(writeConfig(t, dir, minimalLabel))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "`version` is required")

	_, err = Load(writeConfig(t, dir, "version: 99\n"+minimalLabel))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported version 99")
}

func TestValidateRejectsKeyWithNoValues(t *testing.T) {
	dir := t.TempDir()
	_, err := Load(writeConfig(t, dir, "version: 1\nlabels:\n  tier:\n    description: nothing here\n"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), `label key "tier" declares no values`)
}

func TestValidateRejectsEmptyValue(t *testing.T) {
	dir := t.TempDir()
	_, err := Load(writeConfig(t, dir, "version: 1\nlabels:\n  tier:\n    values:\n      hollow:\n        description: nothing\n"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "label tier=hollow defines no activity or workflow attributes")
}

func TestValidateRejectsBadNames(t *testing.T) {
	dir := t.TempDir()

	_, err := Load(writeConfig(t, dir, "version: 1\nlabels:\n  Tier_Name:\n    values:\n      a:\n        activity:\n          max-retries: 1\n"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid label key")

	_, err = Load(writeConfig(t, dir, "version: 1\nlabels:\n  tier:\n    values:\n      Critical:\n        activity:\n          max-retries: 1\n"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), `invalid value "Critical" for label key "tier"`)
}

func TestAnnotationLinesActivity(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(writeConfig(t, dir, validConfig))
	require.NoError(t, err)

	// defaults first, then pairs in source order.
	lines, err := cfg.AnnotationLines("activity", []Pair{{"access", "bulk"}, {"tier", "critical"}})
	require.NoError(t, err)
	assert.Equal(t, []string{
		"// @start-to-close-timeout 1m",
		"// @start-to-close-timeout 1h",
		"// @heartbeat-timeout 1m",
		"// @max-retries 870",
	}, lines)
}

func TestAnnotationLinesWorkflowSkipsActivityOnlyValue(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(writeConfig(t, dir, validConfig))
	require.NoError(t, err)

	// access=read-only has no workflow block, so it contributes nothing here.
	// The defaults block is activity-only too.
	lines, err := cfg.AnnotationLines("workflow", []Pair{{"access", "read-only"}, {"tier", "critical"}})
	require.NoError(t, err)
	assert.Equal(t, []string{
		"// @execution-timeout 24h",
		// memo keys are sorted for deterministic output
		"// @memo owner platform",
		"// @memo tier critical",
	}, lines)
}

func TestAnnotationLinesWorkflowTaskQueueMapsToWorkflowAnnotation(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(writeConfig(t, dir, `
version: 1
labels:
  queue:
    values:
      mine:
        workflow:
          task-queue: my-queue
`))
	require.NoError(t, err)

	lines, err := cfg.AnnotationLines("workflow", []Pair{{"queue", "mine"}})
	require.NoError(t, err)
	assert.Equal(t, []string{"// @workflow-task-queue my-queue"}, lines)
}

func TestAnnotationLinesUnknownKey(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(writeConfig(t, dir, validConfig))
	require.NoError(t, err)

	_, err = cfg.AnnotationLines("activity", []Pair{{"nope", "x"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown label key "nope"`)
	assert.Contains(t, err.Error(), "access, tier")
}

func TestAnnotationLinesUnknownValue(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(writeConfig(t, dir, validConfig))
	require.NoError(t, err)

	_, err = cfg.AnnotationLines("activity", []Pair{{"access", "sideways"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown value "sideways" for label key "access"`)
	assert.Contains(t, err.Error(), "bulk, read-only")
}

func TestAnnotationLinesRejectsDuplicateKey(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(writeConfig(t, dir, validConfig))
	require.NoError(t, err)

	_, err = cfg.AnnotationLines("activity", []Pair{{"access", "read-only"}, {"access", "bulk"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `label key "access" set twice`)
}

func TestAnnotationLinesNilConfig(t *testing.T) {
	var cfg *Config

	lines, err := cfg.AnnotationLines("activity", nil)
	require.NoError(t, err)
	assert.Empty(t, lines)

	_, err = cfg.AnnotationLines("activity", []Pair{{"access", "read-only"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no temporal-gen.yaml was found")
}

func TestAnnotationLinesUnknownKindIsInert(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(writeConfig(t, dir, validConfig))
	require.NoError(t, err)

	// Queries/signals/updates carry no configurable options.
	lines, err := cfg.AnnotationLines("query", []Pair{{"tier", "critical"}})
	require.NoError(t, err)
	assert.Empty(t, lines)
}

func TestDiscoverWalksUp(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte("module x\n"), 0o644))
	writeConfig(t, root, validConfig)

	nested := filepath.Join(root, "a", "b", "c")
	require.NoError(t, os.MkdirAll(nested, 0o755))

	cfg, err := Discover(nested)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, filepath.Join(root, "temporal-gen.yaml"), cfg.Path())
}

func TestDiscoverPrefersNearest(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte("module x\n"), 0o644))
	writeConfig(t, root, validConfig)

	nested := filepath.Join(root, "a")
	require.NoError(t, os.MkdirAll(nested, 0o755))
	writeConfig(t, nested, "version: 1\n"+minimalLabel)

	cfg, err := Discover(nested)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, []string{"tier"}, cfg.Keys())
}

func TestDiscoverStopsAtModuleRoot(t *testing.T) {
	outer := t.TempDir()
	// A config above the module root must not leak into the module.
	writeConfig(t, outer, validConfig)

	module := filepath.Join(outer, "module")
	require.NoError(t, os.MkdirAll(module, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(module, "go.mod"), []byte("module x\n"), 0o644))

	cfg, err := Discover(module)
	require.NoError(t, err)
	assert.Nil(t, cfg)
}

func TestDiscoverNoConfig(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "go.mod"), []byte("module x\n"), 0o644))

	cfg, err := Discover(root)
	require.NoError(t, err)
	assert.Nil(t, cfg)
}
