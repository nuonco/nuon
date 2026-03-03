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

type subdirPermissionsTestStruct struct {
	Obj string `features:"get"`
}

type subdirStackTestStruct struct {
	Template string `features:"get"`
}

type subdirIsolationTestStruct struct {
	Permissions *subdirPermissionsTestStruct
	Stack       *subdirStackTestStruct
}

type subdirPoliciesTestStruct struct {
	Contents string `features:"get"`
}

type subdirFallbackTestStruct struct {
	Policies *subdirPoliciesTestStruct
}

type nestedPolicyTestPolicy struct {
	Contents string `features:"get"`
}

type nestedPolicyTestRole struct {
	Policies []nestedPolicyTestPolicy
}

type nestedPolicyTestPermissions struct {
	Roles []nestedPolicyTestRole
}

type nestedPolicyTestApp struct {
	Permissions *nestedPolicyTestPermissions
}

type sourceAwarePolicyTestStruct struct {
	Contents   string `features:"get"`
	SourceFile string
}

func (s sourceAwarePolicyTestStruct) GetSourceFile() string {
	return s.SourceFile
}

type sourceAwarePoliciesTestApp struct {
	Policies []sourceAwarePolicyTestStruct
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
		// NOTE: The following tests are commented out because they are flaky.
		// They make network calls to private GitHub repos and frequently
		// timeout with the 1-second deadline, causing CI failures.
		//
		// "git_repo_file": {
		// 	input: testStruct{
		// 		Obj: "https://github.com/nuonco/byoc/blob/main/byoc-nuon/policies/set-karpenter-non-cpu-limits.yaml",
		// 	},
		// 	outputFn: func(t *testing.T, ts testStruct) {
		// 		require.NotEqual(t, ts.Obj, "https://github.com/nuonco/byoc/blob/main/byoc-nuon/policies/set-karpenter-non-cpu-limits.yaml")
		// 		require.NotEmpty(t, ts.Obj)
		// 	},
		// },
		// "git_tag_file": {
		// 	input: testStruct{
		// 		Obj: "https://github.com/nuonco/aws-eks-sandbox/blob/0.0.0/README.md",
		// 	},
		// 	outputFn: func(t *testing.T, ts testStruct) {
		// 		require.NotEqual(t, ts.Obj, "https://github.com/nuonco/aws-eks-sandbox/blob/0.0.0/README.md")
		// 		require.NotEmpty(t, ts.Obj)
		// 	},
		// },
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

func TestGetAll_DoesNotLeakSubdirAcrossSiblingFields(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "getall-subdir-isolation-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	err = os.MkdirAll(filepath.Join(tmpDir, "permissions"), 0o755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "permissions", "permission.txt"), []byte("permission content"), 0o644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "permissions", "template.yaml"), []byte("permissions template"), 0o644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "template.yaml"), []byte("stack template"), 0o644)
	require.NoError(t, err)

	input := subdirIsolationTestStruct{
		Permissions: &subdirPermissionsTestStruct{Obj: "./permission.txt"},
		Stack:       &subdirStackTestStruct{Template: "./template.yaml"},
	}

	ctx := context.Background()
	err = Parse(ctx, &input, &Options{
		FieldTimeout: time.Second,
		RootDir:      tmpDir,
	})
	require.NoError(t, err)
	require.Equal(t, "permission content", input.Permissions.Obj)
	require.Equal(t, "stack template", input.Stack.Template)
}

func TestGetAll_FallsBackToRootWhenNamedSubdirDoesNotExist(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "getall-subdir-fallback-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	err = os.WriteFile(filepath.Join(tmpDir, "policy.rego"), []byte("package main"), 0o644)
	require.NoError(t, err)

	input := subdirFallbackTestStruct{
		Policies: &subdirPoliciesTestStruct{Contents: "./policy.rego"},
	}

	err = Parse(context.Background(), &input, &Options{
		FieldTimeout: time.Second,
		RootDir:      tmpDir,
	})
	require.NoError(t, err)
	require.Equal(t, "package main", input.Policies.Contents)
}

func TestGetAll_PermissionRolePoliciesUsePermissionsSubdir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "getall-permissions-nested-policies-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	err = os.MkdirAll(filepath.Join(tmpDir, "permissions"), 0o755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "permissions", "policy.json"), []byte("policy content"), 0o644)
	require.NoError(t, err)

	input := nestedPolicyTestApp{
		Permissions: &nestedPolicyTestPermissions{
			Roles: []nestedPolicyTestRole{{
				Policies: []nestedPolicyTestPolicy{{Contents: "./policy.json"}},
			}},
		},
	}

	err = Parse(context.Background(), &input, &Options{FieldTimeout: time.Second, RootDir: tmpDir})
	require.NoError(t, err)
	require.Equal(t, "policy content", input.Permissions.Roles[0].Policies[0].Contents)
}

func TestGetAll_UsesPolicySourceFileDirForRelativePaths(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "getall-policy-source-dir-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	err = os.MkdirAll(filepath.Join(tmpDir, "policies"), 0o755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(tmpDir, "policies", "disallow-ingress-nginx-custom-snippets.yml"), []byte("apiVersion: kyverno.io/v1"), 0o644)
	require.NoError(t, err)

	input := sourceAwarePoliciesTestApp{
		Policies: []sourceAwarePolicyTestStruct{{
			Contents:   "./disallow-ingress-nginx-custom-snippets.yml",
			SourceFile: "policies/nginx_disallow.toml",
		}},
	}

	err = Parse(context.Background(), &input, &Options{FieldTimeout: time.Second, RootDir: tmpDir})
	require.NoError(t, err)
	require.Equal(t, "apiVersion: kyverno.io/v1", input.Policies[0].Contents)
}
