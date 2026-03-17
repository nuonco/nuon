package main_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/nuonco/nuon/pkg/gen/temporal-gen-v2/cmd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "runtime.Caller failed")
	return filepath.Dir(thisFile)
}

func cleanupGenFiles(t *testing.T, dirs ...string) {
	t.Helper()
	t.Cleanup(func() {
		for _, d := range dirs {
			err := filepath.Walk(d, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if !info.IsDir() && len(path) > 7 && path[len(path)-7:] == "_gen.go" {
					os.Remove(path)
				}
				return nil
			})
			if err != nil {
				t.Logf("cleanup walk error: %v", err)
			}
		}
	})
}

func TestGenerateExamples(t *testing.T) {
	root := repoRoot(t)
	examplesDir := filepath.Join(root, "examples")
	cleanupGenFiles(t, examplesDir)

	rootCmd := cmd.NewRootCmd()
	rootCmd.SetArgs([]string{"generate", "--validate", "--recursive", examplesDir})
	require.NoError(t, rootCmd.Execute())

	// Verify activity_gen.go
	activityGen := filepath.Join(examplesDir, "activity_gen.go")
	require.FileExists(t, activityGen)
	content, err := os.ReadFile(activityGen)
	require.NoError(t, err)
	assert.Contains(t, string(content), "func AwaitSimpleActivity(")
	assert.Contains(t, string(content), "func AwaitComplexActivity(")
	assert.Contains(t, string(content), "THIS FILE IS GENERATED. DO NOT EDIT.")

	// Verify workflow_gen.go
	workflowGen := filepath.Join(examplesDir, "workflow_gen.go")
	require.FileExists(t, workflowGen)
	wfContent, err := os.ReadFile(workflowGen)
	require.NoError(t, err)
	assert.Contains(t, string(wfContent), "func AwaitSimpleWorkflow(")

	// Verify recursive: subdir was processed
	subdirGen := filepath.Join(examplesDir, "subdir", "subdir_activity_gen.go")
	require.FileExists(t, subdirGen)
	subdirContent, err := os.ReadFile(subdirGen)
	require.NoError(t, err)
	assert.Contains(t, string(subdirContent), "func AwaitSubdirActivity(")
}

func TestGenerateWithDependencies(t *testing.T) {
	root := repoRoot(t)
	testdataDir := filepath.Join(root, "testdata", "deps")
	cleanupGenFiles(t, testdataDir)

	rootCmd := cmd.NewRootCmd()
	rootCmd.SetArgs([]string{"generate", "--validate", "--recursive", testdataDir})
	require.NoError(t, rootCmd.Execute())

	// Both packages should have generated files
	pkgbGen := filepath.Join(testdataDir, "pkgb", "activities_gen.go")
	pkgcGen := filepath.Join(testdataDir, "pkgc", "activities_gen.go")

	require.FileExists(t, pkgcGen, "pkgc (dependency) should be generated")
	pkgcContent, err := os.ReadFile(pkgcGen)
	require.NoError(t, err)
	assert.Contains(t, string(pkgcContent), "func AwaitPkgcActivity(")

	require.FileExists(t, pkgbGen, "pkgb (dependent) should be generated")
	pkgbContent, err := os.ReadFile(pkgbGen)
	require.NoError(t, err)
	assert.Contains(t, string(pkgbContent), "func AwaitPkgbActivity(")
}

func TestParallelMatchesSequential(t *testing.T) {
	root := repoRoot(t)
	examplesDir := filepath.Join(root, "examples")
	cleanupGenFiles(t, examplesDir)

	// Run with parallelism=1 (sequential)
	rootCmd := cmd.NewRootCmd()
	rootCmd.SetArgs([]string{"generate", "--validate", "--recursive", "--parallelism", "1", examplesDir})
	require.NoError(t, rootCmd.Execute())

	// Collect sequential output
	seqFiles := map[string][]byte{}
	filepath.Walk(examplesDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && len(path) > 7 && path[len(path)-7:] == "_gen.go" {
			data, readErr := os.ReadFile(path)
			if readErr == nil {
				seqFiles[path] = data
			}
			os.Remove(path)
		}
		return nil
	})

	// Run with parallelism=4
	rootCmd = cmd.NewRootCmd()
	rootCmd.SetArgs([]string{"generate", "--validate", "--recursive", "--parallelism", "4", examplesDir})
	require.NoError(t, rootCmd.Execute())

	// Compare outputs
	for path, seqContent := range seqFiles {
		parContent, err := os.ReadFile(path)
		require.NoError(t, err, "parallel run should produce %s", path)
		assert.Equal(t, string(seqContent), string(parContent),
			"parallel and sequential output should match for %s", path)
	}
}
