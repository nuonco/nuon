package ignorechanges

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
)

func TestEvaluateRun(t *testing.T) {
	tests := map[string]struct {
		paths       []string
		pathsIgnore []string
		regex       string
		files       []string
		ignored     bool
		status      string
	}{
		"skipped by paths": {
			paths: []string{"apps/acme/**"}, files: []string{"apps/other/main.go"},
			ignored: true, status: pathsStatusDescription,
		},
		"skipped by paths_ignore": {
			pathsIgnore: []string{"docs/**", "**/*.md"}, files: []string{"docs/a.txt", "README.md"},
			ignored: true, status: pathsStatusDescription,
		},
		"paths pass then regex ignores": {
			paths: []string{"apps/acme/**"}, regex: `\.md$`, files: []string{"apps/acme/README.md"},
			ignored: true, status: regexStatusDescription,
		},
		"paths pass and regex misses": {
			paths: []string{"apps/acme/**"}, regex: `\.md$`, files: []string{"apps/acme/main.go"},
		},
		"no filters": {
			files: []string{"apps/acme/main.go"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			filter, err := config.NewRunPathFilter(tt.paths, tt.pathsIgnore)
			require.NoError(t, err)
			decision, status, err := EvaluateRun(filter, tt.regex, tt.files)
			require.NoError(t, err)
			require.Equal(t, tt.ignored, decision.Ignored)
			if tt.ignored {
				require.Equal(t, tt.status, status)
				require.NotEmpty(t, decision.Reason)
			}
		})
	}
}

func TestEvaluate(t *testing.T) {
	tests := map[string]struct {
		pattern string
		paths   []string
		ignored bool
		wantErr bool
	}{
		"disabled":       {paths: []string{"docs/readme.md"}},
		"all match":      {pattern: `^docs/`, paths: []string{"docs/a.md", "docs/b.md"}, ignored: true},
		"one misses":     {pattern: `^docs/`, paths: []string{"docs/a.md", "src/main.go"}},
		"empty paths":    {pattern: `.*`, ignored: true},
		"ignore all":     {pattern: `.*`, paths: []string{"src/main.go"}, ignored: true},
		"invalid regexp": {pattern: `[`, paths: []string{"src/main.go"}, wantErr: true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			decision, err := Evaluate(tt.pattern, tt.paths)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.ignored, decision.Ignored)
		})
	}
}
