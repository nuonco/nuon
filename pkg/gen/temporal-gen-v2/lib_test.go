package main_test

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

	assert.Contains(t, body, "options.StartToCloseTimeout = time.Duration(45000000000)")
	assert.Contains(t, body, "MaximumAttempts: int32(3)")
	assert.NotContains(t, body, "time.Duration(300000000000)")
	assert.NotContains(t, body, "MaximumAttempts: int32(9)")

	assert.Contains(t, body, "options.HeartbeatTimeout = time.Duration(20000000000)")
}

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
