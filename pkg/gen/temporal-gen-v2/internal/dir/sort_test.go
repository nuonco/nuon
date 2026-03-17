package dir_test

import (
	"context"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/internal/dir"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func pkgDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller failed")
	// thisFile is .../internal/dir/sort_test.go
	// go up to pkg/gen/temporal-gen-v2/
	return filepath.Join(filepath.Dir(thisFile), "..", "..")
}

func TestGetDependencyLevels_Examples(t *testing.T) {
	root := pkgDir(t)
	examplesPattern := filepath.Join(root, "examples") + "/..."

	ctx := context.Background()
	levels, err := dir.GetDependencyLevels(ctx, examplesPattern)
	require.NoError(t, err)
	assert.NotEmpty(t, levels, "expected at least one level from examples")

	// Flatten all packages and verify both examples packages are present
	var allPkgs []string
	for _, level := range levels {
		allPkgs = append(allPkgs, level...)
	}
	require.NotEmpty(t, allPkgs)

	hasExamples := false
	hasSubdir := false
	for _, p := range allPkgs {
		if strings.HasSuffix(p, "/examples") {
			hasExamples = true
		}
		if strings.HasSuffix(p, "/subdir") {
			hasSubdir = true
		}
	}
	assert.True(t, hasExamples, "expected examples package in levels")
	assert.True(t, hasSubdir, "expected examples/subdir package in levels")
}

func TestGetDependencyLevels_WithDeps(t *testing.T) {
	root := pkgDir(t)
	testdataPattern := filepath.Join(root, "testdata", "deps") + "/..."

	ctx := context.Background()
	levels, err := dir.GetDependencyLevels(ctx, testdataPattern)
	require.NoError(t, err)
	require.Len(t, levels, 2, "pkgc (no deps) should be level 0, pkgb (imports pkgc) should be level 1")

	// Level 0: pkgc (no dependencies within target set)
	require.Len(t, levels[0], 1)
	assert.True(t, strings.HasSuffix(levels[0][0], "/pkgc"),
		"expected pkgc at level 0, got %s", levels[0][0])

	// Level 1: pkgb (depends on pkgc)
	require.Len(t, levels[1], 1)
	assert.True(t, strings.HasSuffix(levels[1][0], "/pkgb"),
		"expected pkgb at level 1, got %s", levels[1][0])
}

func TestGetDependencyLevels_Deterministic(t *testing.T) {
	root := pkgDir(t)
	examplesPattern := filepath.Join(root, "examples") + "/..."

	ctx := context.Background()
	levels1, err := dir.GetDependencyLevels(ctx, examplesPattern)
	require.NoError(t, err)

	levels2, err := dir.GetDependencyLevels(ctx, examplesPattern)
	require.NoError(t, err)

	require.Equal(t, len(levels1), len(levels2))
	for i := range levels1 {
		assert.Equal(t, levels1[i], levels2[i], "level %d should be deterministic", i)
	}
}
