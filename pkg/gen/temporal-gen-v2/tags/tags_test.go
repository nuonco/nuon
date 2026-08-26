package tags

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/generics"
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
tags:
  db-only:
    start-to-close-timeout: 30s
    max-retries: 3
  bulk:
    start-to-close-timeout: 1h
    heartbeat-timeout: 1m
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
	assert.Equal(t, []string{"bulk", "critical", "db-only"}, cfg.Names())
}

func TestLoadRejectsUnknownAttribute(t *testing.T) {
	dir := t.TempDir()
	// "start-to-close" is a typo for "start-to-close-timeout" and must not be
	// silently ignored.
	path := writeConfig(t, dir, `
version: 1
tags:
  oops:
    start-to-close: 30s
`)

	_, err := Load(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "start-to-close")
}

func TestLoadRejectsStructuralAttribute(t *testing.T) {
	dir := t.TempDir()
	// @as-wrapper is deliberately not settable from a tag.
	path := writeConfig(t, dir, `
version: 1
tags:
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
tags:
  oops:
    start-to-close-timeout: 30 seconds
`)

	_, err := Load(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid duration")
}

const minimalTag = "tags:\n  critical:\n    max-retries: 1\n"

func TestValidateVersion(t *testing.T) {
	dir := t.TempDir()

	_, err := Load(writeConfig(t, dir, minimalTag))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "`version` is required")

	_, err = Load(writeConfig(t, dir, "version: 99\n"+minimalTag))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported version 99")
}

// A config built in code does not have to restate the version.
func TestValidateInCodeVersionOptional(t *testing.T) {
	cfg := &Config{Tags: map[string]*Attrs{"critical": {MaxRetries: generics.ToPtr(870)}}}
	require.NoError(t, cfg.Validate())
	assert.Equal(t, SupportedVersion, cfg.Version)
}

func TestValidateInCodeErrorsSayCode(t *testing.T) {
	cfg := &Config{Tags: map[string]*Attrs{"hollow": {}}}
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), `code: tag "hollow" sets no attributes`)
}

func TestValidateInCodeRejectsBadDuration(t *testing.T) {
	cfg := &Config{Tags: map[string]*Attrs{"slow": {StartToCloseTimeout: "30 seconds"}}}
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid duration")
}

func TestValidateRejectsEmptyTag(t *testing.T) {
	dir := t.TempDir()
	_, err := Load(writeConfig(t, dir, "version: 1\ntags:\n  hollow:\n"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), `tag "hollow" sets no attributes`)
}

func TestValidateRejectsBadNames(t *testing.T) {
	dir := t.TempDir()

	_, err := Load(writeConfig(t, dir, "version: 1\ntags:\n  Db_Only:\n    max-retries: 1\n"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), `invalid tag "Db_Only"`)
}

func TestAnnotationLines(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(writeConfig(t, dir, validConfig))
	require.NoError(t, err)

	// defaults first, then tags in source order.
	lines, err := cfg.AnnotationLines([]string{"bulk", "critical"})
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
tags:
  mine:
    task-queue: my-queue
`))
	require.NoError(t, err)

	lines, err := cfg.AnnotationLines([]string{"mine"})
	require.NoError(t, err)
	assert.Equal(t, []string{"// @task-queue my-queue"}, lines)
}

func TestAnnotationLinesUnknownTag(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(writeConfig(t, dir, validConfig))
	require.NoError(t, err)

	_, err = cfg.AnnotationLines([]string{"nope"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown tag "nope"`)
	assert.Contains(t, err.Error(), "bulk, critical, db-only")
}

func TestAnnotationLinesUnknownTagInCodeConfig(t *testing.T) {
	cfg := &Config{Tags: map[string]*Attrs{"critical": {MaxRetries: generics.ToPtr(870)}}}
	require.NoError(t, cfg.Validate())

	_, err := cfg.AnnotationLines([]string{"nope"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown tag "nope" (declared in code: critical)`)
}

func TestAnnotationLinesRejectsDuplicateTag(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(writeConfig(t, dir, validConfig))
	require.NoError(t, err)

	_, err = cfg.AnnotationLines([]string{"bulk", "bulk"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `tag "bulk" set twice`)
}

func TestAnnotationLinesNilConfig(t *testing.T) {
	var cfg *Config

	lines, err := cfg.AnnotationLines(nil)
	require.NoError(t, err)
	assert.Empty(t, lines)

	_, err = cfg.AnnotationLines([]string{"db-only"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no tag config was found")
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
	writeConfig(t, nested, "version: 1\n"+minimalTag)

	cfg, err := Discover(nested)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, []string{"critical"}, cfg.Names())
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
