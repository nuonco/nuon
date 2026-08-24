package main_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
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
				if !info.IsDir() && strings.HasSuffix(path, "_gen.go") {
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

	rootCmd := cmd.NewRootCmd()
	rootCmd.SetArgs([]string{"generate", "--validate", "--imports", "--recursive", examplesDir})
	require.NoError(t, rootCmd.Execute())

	// Verify activity_gen.go
	activityGen := filepath.Join(examplesDir, "activity_gen.go")
	require.FileExists(t, activityGen)
	content, err := os.ReadFile(activityGen)
	require.NoError(t, err)
	assert.Contains(t, string(content), "func AwaitSimpleActivity(")
	assert.Contains(t, string(content), "func AwaitComplexActivity(")
	assert.Contains(t, string(content), "THIS FILE IS GENERATED. DO NOT EDIT.")

	simpleAwait := string(content[strings.Index(string(content), "func AwaitSimpleActivity("):strings.Index(string(content), "func AwaitComplexActivity(")])
	assert.Contains(t, simpleAwait, "MaximumAttempts: int32(870)")
	assert.NotContains(t, simpleAwait, "options.ScheduleToCloseTimeout = time.Duration(")
	assert.Contains(t, string(content), "options.ScheduleToCloseTimeout = time.Duration(3600000000000)")

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

// funcBody returns the generated source between name and the next top-level
// func, so assertions are scoped to a single Await/Exec wrapper.
func funcBody(t *testing.T, content, name string) string {
	t.Helper()
	start := strings.Index(content, name)
	require.GreaterOrEqual(t, start, 0, "%s not found in generated output", name)
	rest := content[start+len(name):]
	if end := strings.Index(rest, "\nfunc "); end >= 0 {
		return rest[:end]
	}
	return rest
}

func TestGenerateLabels(t *testing.T) {
	root := repoRoot(t)
	examplesDir := filepath.Join(root, "examples")

	rootCmd := cmd.NewRootCmd()
	rootCmd.SetArgs([]string{"generate", "--validate", "--imports", "--recursive", examplesDir})
	require.NoError(t, rootCmd.Execute())

	raw, err := os.ReadFile(filepath.Join(examplesDir, "labels_gen.go"))
	require.NoError(t, err)
	content := string(raw)

	// No labels: the `defaults` block still supplies start-to-close 1m.
	unlabeled := funcBody(t, content, "func AwaitUnlabeledActivity(")
	assert.Contains(t, unlabeled, "options.StartToCloseTimeout = time.Duration(60000000000)")

	// access=read-only: 30s / 3 attempts override the 1m default.
	single := funcBody(t, content, "func AwaitSingleLabelActivity(")
	assert.Contains(t, single, "options.StartToCloseTimeout = time.Duration(30000000000)")
	assert.Contains(t, single, "MaximumAttempts: int32(3)")

	// Two distinct label keys both contribute in full: access=bulk supplies
	// the timeouts, tier=critical supplies the retry count.
	multi := funcBody(t, content, "func AwaitMultiLabelActivity(")
	assert.Contains(t, multi, "options.StartToCloseTimeout = time.Duration(3600000000000)")
	assert.Contains(t, multi, "options.ScheduleToCloseTimeout = time.Duration(21600000000000)")
	assert.Contains(t, multi, "options.HeartbeatTimeout = time.Duration(60000000000)")
	assert.Contains(t, multi, "MaximumAttempts: int32(870)")

	// An explicit annotation beats the label it conflicts with.
	overridden := funcBody(t, content, "func AwaitLabelOverriddenActivity(")
	assert.Contains(t, overridden, "options.StartToCloseTimeout = time.Duration(10000000000)")
	assert.Contains(t, overridden, "MaximumAttempts: int32(3)")

	// A label value carrying both blocks contributes only the matching one.
	crossWorkflow := funcBody(t, content, "func ExecLabeledWorkflow(")
	assert.Contains(t, crossWorkflow, "WorkflowExecutionTimeout: time.Duration(86400000000000)")
	assert.Contains(t, crossWorkflow, `"tier": "critical"`)

	// Applied labels are recorded in the generated doc comment.
	assert.Contains(t, content, "// labels: access=bulk, tier=critical")
}

func TestGenerateUnknownLabelFails(t *testing.T) {
	root := repoRoot(t)
	badDir := filepath.Join(root, "testdata", "badlabel")
	cleanupGenFiles(t, badDir)

	rootCmd := cmd.NewRootCmd()
	rootCmd.SetArgs([]string{"generate", "--validate", badDir})
	err := rootCmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown label key "does-not-exist"`)
}

// A config below the module root must not leak into sibling trees: the deps
// fixture sits next to testdata/badlabel but must generate without labels.
func TestGenerateIgnoresSiblingConfig(t *testing.T) {
	root := repoRoot(t)
	depsDir := filepath.Join(root, "testdata", "deps")
	cleanupGenFiles(t, depsDir)

	rootCmd := cmd.NewRootCmd()
	rootCmd.SetArgs([]string{"generate", "--validate", "--recursive", depsDir})
	require.NoError(t, rootCmd.Execute())
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

	// Run with parallelism=1 (sequential)
	rootCmd := cmd.NewRootCmd()
	rootCmd.SetArgs([]string{"generate", "--validate", "--imports", "--recursive", "--parallelism", "1", examplesDir})
	require.NoError(t, rootCmd.Execute())

	// Collect sequential output
	seqFiles := map[string][]byte{}
	filepath.Walk(examplesDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(path, "_gen.go") {
			data, readErr := os.ReadFile(path)
			if readErr == nil {
				seqFiles[path] = data
			}
		}
		return nil
	})

	// Run with parallelism=4
	rootCmd = cmd.NewRootCmd()
	rootCmd.SetArgs([]string{"generate", "--validate", "--imports", "--recursive", "--parallelism", "4", examplesDir})
	require.NoError(t, rootCmd.Execute())

	// Compare outputs
	for path, seqContent := range seqFiles {
		parContent, err := os.ReadFile(path)
		require.NoError(t, err, "parallel run should produce %s", path)
		assert.Equal(t, string(seqContent), string(parContent),
			"parallel and sequential output should match for %s", path)
	}
}
