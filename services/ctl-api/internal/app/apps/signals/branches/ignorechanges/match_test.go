package ignorechanges

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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
