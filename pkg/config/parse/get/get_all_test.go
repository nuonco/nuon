package get

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

type testStruct struct {
	Obj string `features:"get"`
}

func TestGetAll(t *testing.T) {
	// Create a temporary directory for local file tests
	tmpDir, err := os.MkdirTemp("", "getall-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a test file in the temporary directory
	testFilePath := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFilePath, []byte("test content"), 0o644)
	require.NoError(t, err)

	tests := map[string]struct {
		input testStruct

		outputFn func(*testing.T, testStruct)
	}{
		"abs_file": {
			input: testStruct{
				Obj: "file://" + testFilePath,
			},
			outputFn: func(t *testing.T, ts testStruct) {
				require.Equal(t, "test content", ts.Obj)
			},
		},
		"local_file": {
			input: testStruct{
				Obj: "./test.txt",
			},
			outputFn: func(t *testing.T, ts testStruct) {
				require.Equal(t, "test content", ts.Obj)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			err := Parse(ctx, &tc.input, &Options{
				RootDir: tmpDir,
			})
			require.NoError(t, err)
			tc.outputFn(t, tc.input)
		})
	}
}
