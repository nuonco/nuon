package get

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/sourcefs"
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

// TestGetAll_GitFileContents exercises the git source code path end-to-end
// against a local git repository. It guards the bug where go-getter's
// fetchSubmodules toggles DisableSymlinks on the shared client, causing the
// follow-up copyDir to fail on macOS with "copying of symlinks has been
// disabled" because /var/folders -> /private/var.
func TestGetAll_GitFileContents(t *testing.T) {
	gitBin, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git binary not available; skipping git source test")
	}

	repoDir := t.TempDir()

	runGit := func(args ...string) {
		t.Helper()
		cmd := exec.Command(gitBin, args...)
		cmd.Dir = repoDir
		// Make sure the test does not depend on the host git config.
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@example.com",
			"GIT_CONFIG_GLOBAL=/dev/null",
			"GIT_CONFIG_SYSTEM=/dev/null",
		)
		out, err := cmd.CombinedOutput()
		require.NoErrorf(t, err, "git %v failed: %s", args, string(out))
	}

	runGit("init", "-q", "-b", "main")
	require.NoError(t, os.MkdirAll(filepath.Join(repoDir, "kubernetes"), 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(repoDir, "kubernetes", "policy.rego"),
		[]byte("package main\nallow := true\n"),
		0o644,
	))
	runGit("add", "kubernetes/policy.rego")
	runGit("commit", "-q", "-m", "add policy")

	rootDir := t.TempDir()

	input := subdirPoliciesTestStruct{
		Contents: "git::file://" + repoDir + "//kubernetes/policy.rego",
	}

	err = Parse(context.Background(), &input, &Options{
		FieldTimeout: 30 * time.Second,
		RootDir:      rootDir,
	})
	require.NoError(t, err)
	require.Equal(t, "package main\nallow := true\n", input.Contents)
}

func TestGetAll_GitSourceWithoutFileReturnsError(t *testing.T) {
	rootDir := t.TempDir()

	input := subdirPoliciesTestStruct{
		Contents: "git::https://github.com/example/repo.git",
	}

	err := Parse(context.Background(), &input, &Options{
		FieldTimeout: 5 * time.Second,
		RootDir:      rootDir,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "git source must include a `//path/to/file` reference")
}

// TestGetAll_GitDirectoryPassthrough guards module-address values (e.g. gcp
// curated custom stacks): a git source whose subdir is a directory is left
// untouched instead of being inlined.
func TestGetAll_GitDirectoryPassthrough(t *testing.T) {
	gitBin, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git binary not available; skipping git source test")
	}

	repoDir := t.TempDir()

	runGit := func(args ...string) {
		t.Helper()
		cmd := exec.Command(gitBin, args...)
		cmd.Dir = repoDir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@example.com",
			"GIT_CONFIG_GLOBAL=/dev/null",
			"GIT_CONFIG_SYSTEM=/dev/null",
		)
		out, err := cmd.CombinedOutput()
		require.NoErrorf(t, err, "git %v failed: %s", args, string(out))
	}

	runGit("init", "-q", "-b", "main")
	require.NoError(t, os.MkdirAll(filepath.Join(repoDir, "gcp", "modules", "bucket"), 0o755))
	require.NoError(t, os.WriteFile(
		filepath.Join(repoDir, "gcp", "modules", "bucket", "main.tf"),
		[]byte("resource \"google_storage_bucket\" \"main\" {}\n"),
		0o644,
	))
	runGit("add", "gcp/modules/bucket/main.tf")
	runGit("commit", "-q", "-m", "add module")

	rootDir := t.TempDir()

	src := "git::file://" + repoDir + "//gcp/modules/bucket"
	input := subdirPoliciesTestStruct{
		Contents: src,
	}

	err = Parse(context.Background(), &input, &Options{
		FieldTimeout: 30 * time.Second,
		RootDir:      rootDir,
	})
	require.NoError(t, err)
	require.Equal(t, src, input.Contents)
}

func TestRecordLocalFile(t *testing.T) {
	newGet := func(onLocalFile func(string, []byte) error, root string) *get {
		return &get{opts: &Options{RootDir: root, OnLocalFile: onLocalFile}}
	}

	t.Run("records root-relative path for current-dir reference", func(t *testing.T) {
		root := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(root, "values.yaml"), []byte("replicas: 1"), 0o644))

		var recordedPath string
		var recordedContents []byte
		g := newGet(func(path string, contents []byte) error {
			recordedPath, recordedContents = path, contents
			return nil
		}, root)

		require.NoError(t, g.recordLocalFile("./values.yaml", root))
		require.Equal(t, "values.yaml", recordedPath)
		require.Equal(t, []byte("replicas: 1"), recordedContents)
	})

	t.Run("resolves parent-dir reference to a root-relative path", func(t *testing.T) {
		root := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(root, "sub"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(root, "shared.rego"), []byte("package main"), 0o644))

		var recordedPath string
		g := newGet(func(path string, contents []byte) error {
			recordedPath = path
			return nil
		}, root)

		require.NoError(t, g.recordLocalFile("../shared.rego", filepath.Join(root, "sub")))
		require.Equal(t, "shared.rego", recordedPath)
	})

	t.Run("skips references that escape the config root", func(t *testing.T) {
		root := t.TempDir()
		sub := filepath.Join(root, "sub")
		require.NoError(t, os.MkdirAll(sub, 0o755))
		outside := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(outside, "secret.yaml"), []byte("x"), 0o644))

		var called bool
		g := newGet(func(path string, contents []byte) error {
			called = true
			return nil
		}, root)

		require.NoError(t, g.recordLocalFile("../../secret.yaml", sub))
		require.False(t, called)
	})

	t.Run("skips non-local references", func(t *testing.T) {
		var called bool
		g := newGet(func(path string, contents []byte) error {
			called = true
			return nil
		}, t.TempDir())

		require.NoError(t, g.recordLocalFile("https://example.com/file.yaml", t.TempDir()))
		require.False(t, called)
	})

	t.Run("skips symlinks", func(t *testing.T) {
		root := t.TempDir()
		real := filepath.Join(root, "real.yaml")
		require.NoError(t, os.WriteFile(real, []byte("x"), 0o644))
		link := filepath.Join(root, "link.yaml")
		require.NoError(t, os.Symlink(real, link))

		var called bool
		g := newGet(func(path string, contents []byte) error {
			called = true
			return nil
		}, root)

		require.NoError(t, g.recordLocalFile("./link.yaml", root))
		require.False(t, called)
	})

	t.Run("skips directories", func(t *testing.T) {
		root := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(root, "sub"), 0o755))

		var called bool
		g := newGet(func(path string, contents []byte) error {
			called = true
			return nil
		}, root)

		require.NoError(t, g.recordLocalFile("./sub", root))
		require.False(t, called)
	})

	t.Run("is a no-op without a hook", func(t *testing.T) {
		root := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(root, "values.yaml"), []byte("x"), 0o644))
		g := newGet(nil, root)

		require.NoError(t, g.recordLocalFile("./values.yaml", root))
	})
}

func TestGetAll_SourceUsesBorrowedFilesystem(t *testing.T) {
	for _, backend := range []string{"os", "os-symlink", "memory"} {
		t.Run(backend, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "repo")
			require.NoError(t, os.MkdirAll(filepath.Join(root, "policies"), 0o755))
			require.NoError(t, os.WriteFile(filepath.Join(root, "policy.json"), []byte("host root canary"), 0o644))
			require.NoError(t, os.WriteFile(filepath.Join(root, "policies/policy.json"), []byte("host nested canary"), 0o644))

			var fs afero.Fs = afero.NewOsFs()
			if backend == "memory" {
				fs = afero.NewMemMapFs()
			}
			require.NoError(t, fs.MkdirAll(filepath.Join(root, "policies"), 0o755))
			require.NoError(t, afero.WriteFile(fs, filepath.Join(root, "policy.json"), []byte("source root"), 0o644))
			require.NoError(t, afero.WriteFile(fs, filepath.Join(root, "policies/policy.json"), []byte("source nested"), 0o644))
			source, err := sourcefs.New(fs, root)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, source.Close()) })
			if backend != "memory" {
				require.NoError(t, os.Rename(root, root+"-pinned"))
				replacement := root
				if backend == "os-symlink" {
					replacement = root + "-outside"
				}
				require.NoError(t, os.MkdirAll(filepath.Join(replacement, "policies"), 0o755))
				require.NoError(t, os.WriteFile(filepath.Join(replacement, "policy.json"), []byte("replacement root canary"), 0o644))
				require.NoError(t, os.WriteFile(filepath.Join(replacement, "policies/policy.json"), []byte("replacement nested canary"), 0o644))
				if backend == "os-symlink" {
					require.NoError(t, os.Symlink(replacement, root))
				}
			}

			for _, reference := range []string{"./policy.json", filepath.Join(root, "policy.json"), "file://" + filepath.Join(root, "policy.json"), "file::./policy.json", "file::file://" + filepath.Join(root, "policy.json")} {
				t.Run(reference, func(t *testing.T) {
					var recorded []string
					input := testStruct{Obj: reference}
					require.NoError(t, Parse(context.Background(), &input, &Options{
						RootDir: root, Source: source, FieldTimeout: time.Second,
						OnLocalFile: func(name string, contents []byte) error {
							recorded = append(recorded, name+":"+string(contents))
							contents[0] = 'X'
							return nil
						},
					}))
					require.Equal(t, "source root", input.Obj)
					require.Equal(t, []string{"policy.json:source root"}, recorded)
				})
			}

			input := sourceAwarePoliciesTestApp{Policies: []sourceAwarePolicyTestStruct{{
				Contents: "./policy.json", SourceFile: "policies/config.toml",
			}}}
			require.NoError(t, Parse(context.Background(), &input, &Options{RootDir: root, Source: source}))
			require.Equal(t, "source nested", input.Policies[0].Contents)
			contents, err := source.ReadFile("policy.json")
			require.NoError(t, err)
			require.Equal(t, "source root", string(contents))
			if backend == "memory" {
				require.NoError(t, os.WriteFile(filepath.Join(root, "host-only.json"), []byte("host-only canary"), 0o644))
				input := testStruct{Obj: "./host-only.json"}
				require.Error(t, Parse(context.Background(), &input, &Options{RootDir: root, Source: source}))
			}
		})
	}
}

func TestGetAll_SourceRepositoryBoundary(t *testing.T) {
	for _, backend := range []string{"os", "memory"} {
		t.Run(backend, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "repo")
			var fs afero.Fs = afero.NewOsFs()
			if backend == "memory" {
				fs = afero.NewMemMapFs()
			}
			require.NoError(t, fs.MkdirAll(filepath.Join(root, "installs"), 0o755))
			require.NoError(t, afero.WriteFile(fs, filepath.Join(root, "ui.json"), []byte("repository UI"), 0o644))
			outside := filepath.Join(filepath.Dir(root), "outside.json")
			require.NoError(t, os.WriteFile(outside, []byte("host outside canary"), 0o644))
			require.NoError(t, afero.WriteFile(fs, outside, []byte("source outside canary"), 0o644))
			source, err := sourcefs.New(fs, root)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, source.Close()) })
			opts := &Options{RootDir: filepath.Join(root, "installs"), Source: source}

			for _, reference := range []string{"../ui.json", "file::../ui.json", "file://" + filepath.Join(root, "ui.json")} {
				input := testStruct{Obj: reference}
				require.NoError(t, Parse(context.Background(), &input, opts), reference)
				require.Equal(t, "repository UI", input.Obj)
			}
			for _, reference := range []string{
				"../../outside.json",
				"file::../../outside.json",
				"file://" + outside,
				"file::file://" + outside,
				"file://" + root + "/%2e%2e/outside.json",
				"file::https://example.com" + outside,
				"git::file://" + filepath.Dir(root) + "//outside.json",
				"git::../../outside.json",
			} {
				t.Run(reference, func(t *testing.T) {
					input := testStruct{Obj: reference}
					require.Error(t, Parse(context.Background(), &input, opts))
					require.Equal(t, reference, input.Obj)
				})
			}
		})
	}
}

func TestGetAll_SourceSubdirSelectionDoesNotProbeHost(t *testing.T) {
	root := t.TempDir()
	fs := afero.NewMemMapFs()
	require.NoError(t, fs.MkdirAll(filepath.Join(root, "permissions"), 0o755))
	require.NoError(t, afero.WriteFile(fs, filepath.Join(root, "permissions/policy.json"), []byte("source permissions"), 0o644))
	require.NoError(t, afero.WriteFile(fs, filepath.Join(root, "policy.json"), []byte("source root"), 0o644))
	require.NoError(t, os.MkdirAll(filepath.Join(root, "policies"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "policies/policy.json"), []byte("host canary"), 0o644))
	source, err := sourcefs.New(fs, root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, source.Close()) })

	input := subdirIsolationTestStruct{
		Permissions: &subdirPermissionsTestStruct{Obj: "./policy.json"},
		Stack:       &subdirStackTestStruct{Template: "./policy.json"},
	}
	require.NoError(t, Parse(context.Background(), &input, &Options{RootDir: root, Source: source}))
	require.Equal(t, "source permissions", input.Permissions.Obj)
	require.Equal(t, "source root", input.Stack.Template)

	policies := sourceAwarePoliciesTestApp{Policies: []sourceAwarePolicyTestStruct{{
		Contents: "./policy.json", SourceFile: "policies/config.toml",
	}}}
	require.NoError(t, Parse(context.Background(), &policies, &Options{RootDir: root, Source: source}))
	require.Equal(t, "source root", policies.Policies[0].Contents)
}

func TestGetAll_SourceRejectsSymlinkEscapesAndDirectoryReplacement(t *testing.T) {
	root := filepath.Join(t.TempDir(), "repo")
	require.NoError(t, os.MkdirAll(filepath.Join(root, "permissions"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "permissions/policy.json"), []byte("repository policy"), 0o644))
	outside := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(outside, "policy.json"), []byte("outside canary"), 0o644))
	require.NoError(t, os.Symlink(filepath.Join(outside, "policy.json"), filepath.Join(root, "link.json")))
	source, err := sourcefs.New(afero.NewOsFs(), root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, source.Close()) })
	require.NoError(t, os.Rename(filepath.Join(root, "permissions"), filepath.Join(root, "old-permissions")))
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "permissions")))

	for _, reference := range []string{"./link.json", "./permissions/policy.json", "file://" + filepath.Join(root, "link.json")} {
		input := testStruct{Obj: reference}
		require.Error(t, Parse(context.Background(), &input, &Options{RootDir: root, Source: source}), reference)
		require.Equal(t, reference, input.Obj)
	}
	input := subdirIsolationTestStruct{Permissions: &subdirPermissionsTestStruct{Obj: "./policy.json"}}
	require.Error(t, Parse(context.Background(), &input, &Options{RootDir: root, Source: source}))
	policies := sourceAwarePoliciesTestApp{Policies: []sourceAwarePolicyTestStruct{{
		Contents: "./policy.json", SourceFile: "permissions/config.toml",
	}}}
	require.Error(t, Parse(context.Background(), &policies, &Options{RootDir: root, Source: source}))
}

func TestGetAll_SourcePropagatesRecordingError(t *testing.T) {
	root := t.TempDir()
	fs := afero.NewMemMapFs()
	require.NoError(t, fs.MkdirAll(root, 0o755))
	require.NoError(t, afero.WriteFile(fs, filepath.Join(root, "policy.json"), []byte("source"), 0o644))
	source, err := sourcefs.New(fs, root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, source.Close()) })
	recordErr := errors.New("recording failed")
	opts := &Options{Source: source, OnLocalFile: func(string, []byte) error { return recordErr }}
	input := testStruct{Obj: "./policy.json"}
	require.ErrorIs(t, Parse(context.Background(), &input, opts), recordErr)
	require.Empty(t, opts.RootDir)
}

func TestGetAll_SourceRejectsPrivateHTTPAndLocalChecksums(t *testing.T) {
	var requests atomic.Int64
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		_, _ = w.Write([]byte("remote contents"))
	}))
	defer server.Close()
	root := t.TempDir()
	source, err := sourcefs.New(afero.NewOsFs(), root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, source.Close()) })
	opts := &Options{RootDir: root, Source: source, FieldTimeout: 5 * time.Second}
	input := testStruct{Obj: server.URL + "/policy.json"}
	require.ErrorContains(t, Parse(context.Background(), &input, opts), "non-public address")
	input = testStruct{Obj: server.URL + "/policy.json?checksum=" + url.QueryEscape("file:"+server.URL+"/SHA256SUMS")}
	require.ErrorContains(t, Parse(context.Background(), &input, opts), "non-public address")

	for _, checksum := range []string{"./checksum", "file:///outside/checksum", "git::file:///outside//checksum", "file::https://example.com/outside/checksum", server.URL + "/SHA256SUMS?checksum=" + url.QueryEscape("file:file:///outside/checksum")} {
		input := testStruct{Obj: server.URL + "/policy.json?checksum=" + url.QueryEscape("file:"+checksum)}
		require.ErrorContains(t, Parse(context.Background(), &input, opts), "checksum must not reference a local file")
	}
	input = testStruct{Obj: "git@localhost:example"}
	require.ErrorContains(t, Parse(context.Background(), &input, opts), "remote field imports must use public HTTP(S)")
	require.Zero(t, requests.Load())
	input = testStruct{Obj: server.URL + "/policy.json"}
	require.NoError(t, Parse(context.Background(), &input, &Options{RootDir: root, FieldTimeout: 5 * time.Second}))
	require.Equal(t, "remote contents", input.Obj)

	for _, backend := range []string{"replaced-os-root", "missing-memory-root"} {
		t.Run(backend, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "repo")
			var fs afero.Fs = afero.NewOsFs()
			if backend == "missing-memory-root" {
				fs = afero.NewMemMapFs()
			}
			require.NoError(t, fs.MkdirAll(root, 0o755))
			source, err := sourcefs.New(fs, root)
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, source.Close()) })
			if backend == "replaced-os-root" {
				require.NoError(t, os.Rename(root, root+"-pinned"))
				require.NoError(t, os.WriteFile(root, []byte("replacement canary"), 0o644))
			}

			input := testStruct{Obj: server.URL + "/policy.json"}
			opts := &Options{RootDir: root, Source: source, FieldTimeout: 5 * time.Second}
			require.ErrorContains(t, Parse(context.Background(), &input, opts), "non-public address")
			entries, err := source.ReadDir(".")
			require.NoError(t, err)
			require.Empty(t, entries)

			input = testStruct{Obj: server.URL + "/policy.json"}
			opts.Source = nil
			require.ErrorContains(t, Parse(context.Background(), &input, opts), "failed to create temp directory")
			if backend == "replaced-os-root" {
				contents, err := os.ReadFile(root)
				require.NoError(t, err)
				require.Equal(t, "replacement canary", string(contents))
			} else {
				_, err := os.Stat(root)
				require.True(t, os.IsNotExist(err))
			}
		})
	}
}

func TestGetAll_SourceRejectsGitBeforeExecutingIt(t *testing.T) {
	bin := t.TempDir()
	marker := filepath.Join(bin, "executed")
	require.NoError(t, os.WriteFile(filepath.Join(bin, "git"), []byte("#!/bin/sh\nprintf invoked > '"+marker+"'\nexit 1\n"), 0o700))
	t.Setenv("PATH", bin)
	root := t.TempDir()
	source, err := sourcefs.New(afero.NewOsFs(), root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, source.Close()) })
	opts := &Options{RootDir: root, Source: source, FieldTimeout: 10 * time.Second}
	for _, reference := range []string{
		"git::https://example.com/acme/repo.git//policy.json",
		"git::http://127.0.0.1/repo.git//policy.json",
		"git::file:///outside/repo//policy.json",
		"git@localhost:acme/repo//policy.json",
		"git::ext::helper//policy.json",
		"https://example.com/policy.json?checksum=" + url.QueryEscape("file:git::https://example.com/acme/repo.git//checksum"),
	} {
		input := testStruct{Obj: reference}
		require.Error(t, Parse(context.Background(), &input, opts), reference)
	}
	_, err = os.Stat(marker)
	require.True(t, os.IsNotExist(err), "a denied import must not invoke git")
}
