package hooks

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSignalsWithLogStreamsImplementInterface scans all signal packages for
// log-stream creation patterns and asserts that those packages also declare
// a compile-time check for SignalWithLogStream. This prevents new signals
// from creating log streams without opting into the lifecycle-hook cleanup.
func TestSignalsWithLogStreamsImplementInterface(t *testing.T) {
	// Walk upward to find the repo root (contains go.mod).
	root, err := findRepoRoot()
	require.NoError(t, err, "unable to find repo root")

	signalsRoot := filepath.Join(root, "services", "ctl-api", "internal", "app")

	// Patterns that indicate a signal creates or uses a log stream.
	creationPatterns := []string{
		"CreateLogStream",
		"SetLogStreamWorkflowContext",
	}

	// The compile-time assertion we expect in the same package.
	interfaceCheck := "SignalWithLogStream"

	var violations []string

	err = filepath.Walk(signalsRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".go") || strings.HasSuffix(info.Name(), "_test.go") {
			return nil
		}
		// Only look at signal packages (directories under signals/ that
		// contain a signal.go file — skip activity/helper directories).
		if !strings.Contains(path, string(filepath.Separator)+"signals"+string(filepath.Separator)) {
			return nil
		}
		pkgDir := filepath.Dir(path)
		if _, err := os.Stat(filepath.Join(pkgDir, "signal.go")); os.IsNotExist(err) {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := string(data)

		hasCreation := false
		for _, p := range creationPatterns {
			if strings.Contains(content, p) {
				hasCreation = true
				break
			}
		}
		if !hasCreation {
			return nil
		}

		// Check the entire package directory for the interface assertion.
		if !packageContains(pkgDir, interfaceCheck) {
			rel, _ := filepath.Rel(signalsRoot, path)
			violations = append(violations, rel)
		}

		return nil
	})
	require.NoError(t, err)

	assert.Empty(t, violations,
		"The following signal files create log streams but their package does not implement SignalWithLogStream.\n"+
			"Add `var _ signal.SignalWithLogStream = (*Signal)(nil)` and a `LogStreamID() string` method:\n  %s",
		strings.Join(violations, "\n  "))
}

// packageContains checks whether any non-test .go file in dir contains substr.
func packageContains(dir, substr string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		if strings.Contains(string(data), substr) {
			return true
		}
	}
	return false
}

// findRepoRoot walks up from the current working directory to find go.mod.
func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
