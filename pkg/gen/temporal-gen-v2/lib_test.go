package main_test

// Exercises lib.Generate directly, the way services/ctl-api/cmd/gen invokes it.
// Everything in generate_test.go goes through the cobra CLI, which cannot reach
// Options.Tags at all.

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	temporalgen "github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/lib"
	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/tags"
	"github.com/nuonco/nuon/pkg/generics"
)

// libTagsFixture returns the fixture dir and registers cleanup of its generated
// files. Its temporal-gen.yaml declares db-read as 5m / 9 attempts, which the
// in-code configs below deliberately disagree with.
func libTagsFixture(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(repoRoot(t), "testdata", "libtags")
	cleanupGenFiles(t, dir)
	return dir
}

func genFileContent(t *testing.T, dir, name string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, name))
	require.NoError(t, err)
	return string(raw)
}

func TestLibGenerateWithInCodeTags(t *testing.T) {
	dir := libTagsFixture(t)

	err := temporalgen.Generate(context.Background(), temporalgen.Options{
		Dir:      dir,
		Validate: true,
		Tags: &tags.Config{
			Defaults: &tags.Attrs{HeartbeatTimeout: "20s"},
			Tags: map[string]*tags.Attrs{
				"db-read": {
					StartToCloseTimeout: "45s",
					MaxRetries:          generics.ToPtr(3),
				},
			},
		},
	})
	require.NoError(t, err)

	body := funcBody(t, genFileContent(t, dir, "activities_gen.go"), "func AwaitTaggedActivity(")

	// The in-code config wins outright: 45s / 3 attempts, not the 5m / 9 in the
	// temporal-gen.yaml sitting next to the fixture.
	assert.Contains(t, body, "options.StartToCloseTimeout = time.Duration(45000000000)")
	assert.Contains(t, body, "MaximumAttempts: int32(3)")
	assert.NotContains(t, body, "time.Duration(300000000000)")
	assert.NotContains(t, body, "MaximumAttempts: int32(9)")

	// The in-code `defaults` block applies too.
	assert.Contains(t, body, "options.HeartbeatTimeout = time.Duration(20000000000)")
}

// Options.Tags is the whole vocabulary, so a tag declared only in a discoverable
// temporal-gen.yaml must still be rejected. Otherwise a stray config file could
// silently widen what an in-code caller accepts.
func TestLibGenerateInCodeTagsSkipDiscovery(t *testing.T) {
	dir := libTagsFixture(t)

	err := temporalgen.Generate(context.Background(), temporalgen.Options{
		Dir:      dir,
		Validate: true,
		Tags: &tags.Config{
			Tags: map[string]*tags.Attrs{
				"something-else": {StartToCloseTimeout: "45s"},
			},
		},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown tag "db-read"`)
	assert.Contains(t, err.Error(), "declared in code: something-else")
}

// Same as above but without Validate, since an unresolvable tag is fatal
// regardless of strict mode.
func TestLibGenerateUnknownInCodeTagFailsWithoutValidate(t *testing.T) {
	dir := libTagsFixture(t)

	err := temporalgen.Generate(context.Background(), temporalgen.Options{
		Dir: dir,
		Tags: &tags.Config{
			Tags: map[string]*tags.Attrs{
				"something-else": {StartToCloseTimeout: "45s"},
			},
		},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown tag "db-read"`)
}

// An invalid in-code config must fail before any file is touched, rather than
// being reported per-activity later.
func TestLibGenerateRejectsInvalidInCodeTags(t *testing.T) {
	dir := libTagsFixture(t)

	err := temporalgen.Generate(context.Background(), temporalgen.Options{
		Dir:      dir,
		Validate: true,
		Tags: &tags.Config{
			Tags: map[string]*tags.Attrs{
				"db-read": {StartToCloseTimeout: "not-a-duration"},
			},
		},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), `code: tag "db-read"`)
	assert.Contains(t, err.Error(), `invalid duration "not-a-duration"`)
	assert.NoFileExists(t, filepath.Join(dir, "activities_gen.go"))
}

// With no Tags set, lib.Generate falls back to discovering the fixture's
// temporal-gen.yaml — the CLI's behaviour, reached through the library.
func TestLibGenerateFallsBackToDiscoveredConfig(t *testing.T) {
	dir := libTagsFixture(t)

	err := temporalgen.Generate(context.Background(), temporalgen.Options{
		Dir:      dir,
		Validate: true,
	})
	require.NoError(t, err)

	body := funcBody(t, genFileContent(t, dir, "activities_gen.go"), "func AwaitTaggedActivity(")
	assert.Contains(t, body, "options.StartToCloseTimeout = time.Duration(300000000000)")
	assert.Contains(t, body, "MaximumAttempts: int32(9)")
}

// NoConfig turns tags off entirely, so an annotated @tag has nothing to resolve
// against and must fail rather than generate a wrapper with no defaults.
func TestLibGenerateNoConfigRejectsTaggedActivity(t *testing.T) {
	dir := libTagsFixture(t)

	err := temporalgen.Generate(context.Background(), temporalgen.Options{
		Dir:      dir,
		Validate: true,
		NoConfig: true,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no tag config was found")
}
