package get

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

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
		"git_repo_file": {
			input: testStruct{
				Obj: "https://github.com/tryretool/byoc-kyverno-policies/blob/main/no-create-role.yaml",
			},
			outputFn: func(t *testing.T, ts testStruct) {
				require.NotEqual(t, ts.Obj, "https://github.com/tryretool/byoc-kyverno-policies/blob/main/no-create-role.yaml")
				require.NotEmpty(t, ts.Obj)
			},
		},
		"git_tag_file": {
			input: testStruct{
				Obj: "https://github.com/nuonco/aws-eks-sandbox/blob/0.0.0/README.md",
			},
			outputFn: func(t *testing.T, ts testStruct) {
				require.NotEqual(t, ts.Obj, "https://github.com/nuonco/aws-eks-sandbox/blob/0.0.0/README.md")
				require.NotEmpty(t, ts.Obj)
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			err := Parse(ctx, &tc.input, &Options{
				FieldTimeout: time.Second,
				RootDir:      tmpDir,
			})
			require.NoError(t, err)
			tc.outputFn(t, tc.input)
		})
	}
}
