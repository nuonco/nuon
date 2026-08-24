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
  start-to-close-timeout: 1m
labels:
  access:
    description: how the activity reaches its data
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
`

func TestLoadValid(t *testing.T) {
	dir := t.TempDir()
	path := writeConfig(t, dir, validConfig)

	cfg, err := Load(path)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, path, cfg.Path())
	assert.Equal(t, []string{"access", "tier"}, cfg.Keys())
	assert.Equal(t, []string{"bulk", "db-only"}, cfg.Values("access"))
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
        start-to-close-timeout: 30 seconds
`)

	_, err := Load(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid duration")
}

const minimalLabel = "labels:\n  tier:\n    values:\n      critical:\n        max-retries: 1\n"

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
	assert.Contains(t, err.Error(), "label tier=hollow sets no attributes")
}

func TestValidateRejectsBadNames(t *testing.T) {
	dir := t.TempDir()

	_, err := Load(writeConfig(t, dir, "version: 1\nlabels:\n  Tier_Name:\n    values:\n      a:\n        max-retries: 1\n"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid label key")

	_, err = Load(writeConfig(t, dir, "version: 1\nlabels:\n  tier:\n    values:\n      Critical:\n        max-retries: 1\n"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), `invalid value "Critical" for label key "tier"`)
}

func TestAnnotationLines(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(writeConfig(t, dir, validConfig))
	require.NoError(t, err)

	// defaults first, then pairs in source order.
	lines, err := cfg.AnnotationLines([]Pair{{"access", "bulk"}, {"tier", "critical"}})
	require.NoError(t, err)
	assert.Equal(t, []string{
		"// @start-to-close-timeout 1m",
		"// @start-to-close-timeout 1h",
		"// @heartbeat-timeout 1m",
		"// @max-retries 870",
	}, lines)
}

func TestAnnotationLinesTaskQueue(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(writeConfig(t, dir, `
version: 1
labels:
  queue:
    values:
      mine:
        task-queue: my-queue
`))
	require.NoError(t, err)

	lines, err := cfg.AnnotationLines([]Pair{{"queue", "mine"}})
	require.NoError(t, err)
	assert.Equal(t, []string{"// @task-queue my-queue"}, lines)
}

func TestAnnotationLinesUnknownKey(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(writeConfig(t, dir, validConfig))
	require.NoError(t, err)

	_, err = cfg.AnnotationLines([]Pair{{"nope", "x"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown label key "nope"`)
	assert.Contains(t, err.Error(), "access, tier")
}

func TestAnnotationLinesUnknownValue(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(writeConfig(t, dir, validConfig))
	require.NoError(t, err)

	_, err = cfg.AnnotationLines([]Pair{{"access", "sideways"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown value "sideways" for label key "access"`)
	assert.Contains(t, err.Error(), "bulk, db-only")
}

func TestAnnotationLinesRejectsDuplicateKey(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(writeConfig(t, dir, validConfig))
	require.NoError(t, err)

	_, err = cfg.AnnotationLines([]Pair{{"access", "db-only"}, {"access", "bulk"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `label key "access" set twice`)
}

func TestAnnotationLinesNilConfig(t *testing.T) {
	var cfg *Config

	lines, err := cfg.AnnotationLines(nil)
	require.NoError(t, err)
	assert.Empty(t, lines)

	_, err = cfg.AnnotationLines([]Pair{{"access", "db-only"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no temporal-gen.yaml was found")
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
