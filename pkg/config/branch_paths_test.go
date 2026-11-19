package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateRunPaths(t *testing.T) {
	tests := []struct {
		name        string
		paths       []string
		pathsIgnore []string
		wantErr     string
	}{
		{name: "empty"},
		{name: "valid paths", paths: []string{"apps/acme/**", "./shared/*.toml", "**/*.md"}},
		{name: "valid paths_ignore", pathsIgnore: []string{"docs/**", "**/*.md"}},
		{name: "mutually exclusive", paths: []string{"a/**"}, pathsIgnore: []string{"b/**"}, wantErr: "mutually exclusive"},
		{name: "traversal", paths: []string{"../secrets/**"}, wantErr: "cannot contain '..'"},
		{name: "nested traversal", paths: []string{"apps/../../etc"}, wantErr: "cannot contain '..'"},
		{name: "negated paths", paths: []string{"apps/**", "!apps/docs/**"}, wantErr: "cannot be negated"},
		{name: "absolute", paths: []string{"/etc/passwd"}, wantErr: "relative to the repository root"},
		{name: "backslash", paths: []string{`apps\acme`}, wantErr: "relative to the repository root"},
		{name: "drive letter", pathsIgnore: []string{"C:/x"}, wantErr: "relative to the repository root"},
		{name: "url", paths: []string{"https://example.com/x"}, wantErr: "relative to the repository root"},
		{name: "control char", paths: []string{"apps/\x00acme"}, wantErr: "control characters"},
		{name: "newline", paths: []string{"apps/acme\n"}, wantErr: "control characters"},
		{name: "empty pattern", paths: []string{"  "}, wantErr: "cannot be empty"},
		{name: "bare negation", paths: []string{"!"}, wantErr: "cannot be negated"},
		{name: "negated paths_ignore", pathsIgnore: []string{"!docs/**"}, wantErr: "cannot be negated"},
		{name: "invalid glob", paths: []string{"apps/[acme"}, wantErr: "invalid pattern"},
		{name: "too long", paths: []string{strings.Repeat("a", maxRunPathLength+1)}, wantErr: "exceeds"},
		{name: "too many double stars", paths: []string{"**/a/**/b/**/c/**/d"}, wantErr: "'**/' segments"},
		{name: "too many patterns", paths: make([]string, maxRunPathPatterns+1), wantErr: "at most"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRunPaths(tt.paths, tt.pathsIgnore)
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tt.wantErr)
		})
	}
}

func TestRunPathFilterMatch(t *testing.T) {
	tests := []struct {
		name        string
		paths       []string
		pathsIgnore []string
		files       []string
		wantRun     bool
	}{
		{name: "no filter runs everything", files: []string{"anything.txt"}, wantRun: true},
		{name: "paths match", paths: []string{"shared/**"}, files: []string{"shared/helm/values.yaml"}, wantRun: true},
		{name: "paths no match", paths: []string{"apps/acme/**"}, files: []string{"README.md"}, wantRun: false},
		{name: "double star matches root files", paths: []string{"**/*.toml"}, files: []string{"metadata.toml"}, wantRun: true},
		{name: "middle double star matches zero dirs", paths: []string{"apps/**/values.yaml"}, files: []string{"apps/values.yaml"}, wantRun: true},
		{name: "middle double star matches nested dirs", paths: []string{"apps/**/values.yaml"}, files: []string{"apps/a/b/values.yaml"}, wantRun: true},
		{name: "single star stays in segment", paths: []string{"apps/*"}, files: []string{"apps/a/b"}, wantRun: false},
		{name: "paths_ignore all ignored", pathsIgnore: []string{"docs/**", "**/*.md"}, files: []string{"docs/a.txt", "apps/README.md"}, wantRun: false},
		{name: "paths_ignore one not ignored", pathsIgnore: []string{"docs/**"}, files: []string{"docs/a.txt", "apps/main.toml"}, wantRun: true},
		{name: "no changed files", paths: []string{"apps/**"}, files: nil, wantRun: false},
		{name: "absolute changed path fails open", paths: []string{"apps/**"}, files: []string{"/etc/passwd"}, wantRun: true},
		{name: "traversal changed path fails open", paths: []string{"apps/**"}, files: []string{"../x"}, wantRun: true},
		{name: "cleaned traversal changed path fails open", paths: []string{"apps/**"}, files: []string{"apps/../../x"}, wantRun: true},
		{name: "control char changed path fails open", pathsIgnore: []string{"**"}, files: []string{"a\nb"}, wantRun: true},
		{name: "empty changed path fails open", pathsIgnore: []string{"**"}, files: []string{""}, wantRun: true},
		{name: "cleaned changed path matches", paths: []string{"apps/**"}, files: []string{"./apps/x/../y"}, wantRun: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := NewRunPathFilter(tt.paths, tt.pathsIgnore)
			require.NoError(t, err)
			d := f.Match(tt.files)
			require.Equal(t, tt.wantRun, d.Run)
			if !d.Run {
				require.NotEmpty(t, d.Reason)
				for _, file := range tt.files {
					require.NotContains(t, d.Reason, file)
				}
			}
		})
	}
}
