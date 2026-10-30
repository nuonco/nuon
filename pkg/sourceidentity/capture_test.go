package sourceidentity_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"hegel.dev/go/hegel"
	"helm.sh/helm/v4/pkg/chart/v2/loader"

	"github.com/nuonco/nuon/pkg/sourcefs"
	"github.com/nuonco/nuon/pkg/sourceidentity"
)

func rootedSource(t testing.TB, fs afero.Fs, root string) *sourcefs.FS {
	source, err := sourcefs.New(fs, root)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, source.Close()) })
	return source
}

type diskFile struct {
	path, data string
	mode       os.FileMode
}

func writeDiskFile(t require.TestingT, fs afero.Fs, root string, file diskFile) {
	name := filepath.Join(root, filepath.FromSlash(file.path))
	require.NoError(t, fs.MkdirAll(filepath.Dir(name), 0755))
	require.NoError(t, afero.WriteFile(fs, name, []byte(file.data), 0600))
	require.NoError(t, fs.Chmod(name, file.mode))
}

func writeTree(t require.TestingT, fs afero.Fs, repo, scope string, files []diskFile) {
	root := filepath.Join(repo, scope)
	require.NoError(t, fs.MkdirAll(root, 0755))
	for _, file := range files {
		writeDiskFile(t, fs, root, file)
	}
}

func modelManifest(files []diskFile) []sourceidentity.File {
	var manifest []sourceidentity.File
	for _, file := range files {
		if slices.Contains(strings.Split(file.path, "/"), ".git") {
			continue
		}
		manifest = append(manifest, sourceidentity.File{Path: file.path, SHA256: sum256(file.data), Executable: file.mode&0111 != 0})
	}
	slices.SortFunc(manifest, func(a, b sourceidentity.File) int { return strings.Compare(a.Path, b.Path) })
	return manifest
}

func checkCapture(t require.TestingT, actual, expected *sourceidentity.Identity, files []diskFile, reasons []string) {
	want := cloneIdentity(expected)
	want.Files, want.Captured = modelManifest(files), true
	want.Reasons = slices.Clone(reasons)
	slices.Sort(want.Reasons)
	want.Reasons = slices.Compact(want.Reasons)
	want.Coverage = "complete"
	if len(want.Reasons) > 0 {
		want.Coverage = "unknown"
	}
	want.Checksum = sum256(identityPreimage(want))
	require.Equal(t, want, actual)
	require.NoError(t, actual.Validate())
	wantEligible := ""
	if len(want.Reasons) == 0 {
		wantEligible = want.Checksum
	}
	require.Equal(t, wantEligible, actual.EligibleChecksum())
}

func temporaryRepo(t *hegel.T) (string, func()) {
	root, err := filepath.EvalSymlinks(os.TempDir())
	require.NoError(t, err)
	dir, err := os.MkdirTemp(root, "sourceidentity-")
	require.NoError(t, err)
	return dir, func() { require.NoError(t, os.RemoveAll(dir)) }
}

func drawTree(t *hegel.T, min int) []diskFile {
	count := hegel.Draw(t, hegel.Integers(min, 5))
	var files []diskFile
	for n := 0; n < count; n++ {
		files = append(files, diskFile{
			path: fmt.Sprintf("dir-%d/é-%d.txt", n%2, n),
			data: hegel.Draw(t, boundedText(32, 64)),
			mode: hegel.Draw(t, hegel.SampledFrom([]os.FileMode{0600, 0644, 0755, 0710, 0601})),
		})
	}
	return files
}

func TestCaptureTreeProperties(t *testing.T) {
	t.Parallel()
	hegel.Test(t, func(t *hegel.T) {
		fs, parent := afero.NewMemMapFs(), "/repositories"
		repo, other := filepath.Join(parent, "checkout-a"), filepath.Join(parent, "checkout-b")
		scope := hegel.Draw(t, hegel.SampledFrom([]string{"", "src", "nested/src"}))
		files := drawTree(t, 0)
		writeTree(t, fs, repo, scope, files)
		reversed := slices.Clone(files)
		slices.Reverse(reversed)
		writeTree(t, fs, other, scope, reversed)
		writeDiskFile(t, fs, filepath.Join(repo, scope), diskFile{".git/HEAD", "metadata-a", 0644})
		writeDiskFile(t, fs, filepath.Join(other, scope), diskFile{".git/HEAD", "metadata-b", 0644})
		writeDiskFile(t, fs, parent, diskFile{"outside.txt", "unrelated", 0644})
		if scope != "" {
			writeDiskFile(t, fs, repo, diskFile{"sibling.txt", "unrelated-a", 0644})
			writeDiskFile(t, fs, other, diskFile{"sibling.txt", "unrelated-b", 0644})
		}
		expected := sourceidentity.New("terraform", "config")
		expected.Root = scope
		before := cloneIdentity(expected)
		source, otherSource := rootedSource(t, fs, repo), rootedSource(t, fs, other)
		actual, err := sourceidentity.Capture(source, expected)
		require.NoError(t, err)
		checkCapture(t, actual, expected, files, nil)
		for _, file := range files {
			name := filepath.Join(other, scope, filepath.FromSlash(file.path))
			require.NoError(t, fs.Chtimes(name, time.Unix(1, 0), time.Unix(2, 0)))
		}
		relocated, err := sourceidentity.Capture(otherSource, expected)
		require.NoError(t, err)
		checkCapture(t, relocated, expected, files, nil)
		require.NoError(t, sourceidentity.Verify(source, actual))
		require.NoError(t, sourceidentity.Verify(otherSource, actual))
		disk, cleanup := temporaryRepo(t)
		defer cleanup()
		writeTree(t, afero.NewOsFs(), disk, scope, files)
		capturedOnDisk, err := sourceidentity.Capture(rootedSource(t, afero.NewOsFs(), disk), expected)
		require.NoError(t, err)
		require.Equal(t, actual, capturedOnDisk)
		require.Equal(t, before, expected)
	})
}

func TestCaptureDriftProperties(t *testing.T) {
	t.Parallel()
	hegel.Test(t, func(t *hegel.T) {
		fs, repo := afero.NewMemMapFs(), "/repo"
		original := drawTree(t, 1)
		current := slices.Clone(original)
		writeTree(t, fs, repo, "src", current)
		expected := sourceidentity.New("terraform", "config")
		expected.Root = "src"
		source := rootedSource(t, fs, repo)
		prepared, err := sourceidentity.Capture(source, expected)
		require.NoError(t, err)
		checkCapture(t, prepared, expected, original, nil)
		before := cloneIdentity(prepared)
		steps := hegel.Draw(t, hegel.Integers(2, 8))
		for step := 0; step < steps; step++ {
			action := hegel.Draw(t, hegel.SampledFrom([]string{"bytes", "execute", "permissions", "add", "remove", "rename", "restore", "timestamp", "outside", "git"}))
			if len(current) == 0 {
				action = "restore"
			}
			switch action {
			case "bytes":
				current[0].data += "!"
				writeDiskFile(t, fs, filepath.Join(repo, "src"), current[0])
			case "execute":
				current[0].mode ^= 0100
				require.NoError(t, fs.Chmod(filepath.Join(repo, "src", current[0].path), current[0].mode))
			case "permissions":
				current[0].mode ^= 0040
				require.NoError(t, fs.Chmod(filepath.Join(repo, "src", current[0].path), current[0].mode))
			case "add":
				current = append(current, diskFile{fmt.Sprintf("new-%d.txt", step), "", 0644})
				writeDiskFile(t, fs, filepath.Join(repo, "src"), current[len(current)-1])
			case "remove":
				require.NoError(t, fs.Remove(filepath.Join(repo, "src", current[0].path)))
				current = current[1:]
			case "rename":
				name := fmt.Sprintf("renamed-%d.txt", step)
				require.NoError(t, fs.Rename(filepath.Join(repo, "src", current[0].path), filepath.Join(repo, "src", name)))
				current[0].path = name
			case "restore":
				require.NoError(t, fs.RemoveAll(filepath.Join(repo, "src")))
				current = slices.Clone(original)
				writeTree(t, fs, repo, "src", current)
			case "timestamp":
				name := filepath.Join(repo, "src", current[0].path)
				require.NoError(t, fs.Chtimes(name, time.Unix(1, 0), time.Unix(2, 0)))
			case "outside":
				writeDiskFile(t, fs, repo, diskFile{"sibling.txt", fmt.Sprint(step), 0644})
			case "git":
				writeDiskFile(t, fs, filepath.Join(repo, "src"), diskFile{".git/HEAD", fmt.Sprint(step), 0644})
			}
			actual, err := sourceidentity.Capture(source, expected)
			require.NoError(t, err)
			checkCapture(t, actual, expected, current, nil)
			drift := !slices.Equal(modelManifest(original), modelManifest(current))
			require.Equal(t, drift, sourceidentity.Verify(source, prepared) != nil, "step %d: %s", step, action)
			require.Equal(t, before, prepared)
		}
	})
}

type closureFixture struct {
	name, kind, path, data string
	extra                  []diskFile
	reasons                []string
}

var closureFixtures = []closureFixture{
	{"tf-output", "terraform", "main.tf", `output "x" { value = 1 }`, nil, nil},
	{"tf-local-module", "terraform", "main.tf", `module "x" { source = "./child" }`, []diskFile{{"child/main.tf", `output "x" { value = 1 }`, 0644}}, nil},
	{"tf-git-module", "terraform", "main.tf", `module "x" { source = "./.git/child" }`, []diskFile{{".git/child/main.tf", `output "x" { value = 1 }`, 0644}}, []string{"terraform_module_outside_captured_source"}},
	{"tf-nested-git-module", "terraform", "main.tf", `module "x" { source = "./nested/.git/child" }`, []diskFile{{"nested/.git/child/main.tf", `output "x" { value = 1 }`, 0644}}, []string{"terraform_module_outside_captured_source"}},
	{"tf-remote-module", "terraform", "main.tf", `module "x" { source = "example.com/acme/module" }`, nil, []string{"terraform_module_outside_captured_source"}},
	{"tf-outside-module", "terraform", "main.tf", `module "x" { source = "../../outside" }`, nil, []string{"terraform_module_outside_captured_source"}},
	{"tf-dynamic-module", "terraform", "main.tf", `module "x" { source = var.source }`, nil, []string{"terraform_module_source_unknown"}},
	{"tf-missing-source", "terraform", "main.tf", `module "x" {}`, nil, []string{"terraform_module_source_unknown"}},
	{"tf-malformed", "terraform", "main.tf", `module "x" {`, nil, []string{"terraform_dependencies_unparseable"}},
	{"tf-filesystem", "terraform", "main.tf", `output "x" { value = file("outside.txt") }`, nil, []string{"terraform_filesystem_expression_uninspected"}},
	{"tf-unlocked", "terraform", "main.tf", `resource "example_thing" "x" {}`, nil, []string{"terraform_provider_closure_uninspected", "terraform_providers_not_locked"}},
	{"tf-locked", "terraform", "main.tf", `provider "example" {}`, []diskFile{{".terraform.lock.hcl", "# lock", 0644}}, []string{"terraform_provider_closure_uninspected"}},
	{"tf-data", "terraform", "main.tf", `data "example_thing" "x" {}`, nil, []string{"terraform_provider_closure_uninspected", "terraform_providers_not_locked", "terraform_runtime_inputs_uninspected"}},
	{"tf-json", "terraform", "main.tf.json", `{}`, nil, []string{"terraform_json_dependencies_uninspected"}},
	{"helm-local", "helm", "Chart.yaml", "apiVersion: v2\nname: acme\nversion: 1.0.0\ndependencies: []\n", nil, nil},
	{"helm-dependency", "helm", "Chart.yaml", "dependencies:\n - name: child\n   version: 1.0.0\n   repository: file://child\n", nil, []string{"helm_dependencies_resolved_at_build"}},
	{"helm-requirements", "helm", "requirements.yaml", "dependencies: [{name: child}]\n", nil, []string{"helm_dependencies_resolved_at_build"}},
	{"helm-malformed", "helm", "Chart.yaml", "dependencies: [", nil, []string{"helm_dependencies_unparseable"}},
	{"helm-values-control", "helm", "values.yaml", "dependencies: [{name: child}]\n", nil, nil},
	{"kustomize-local", "kustomize", "kustomization.yaml", "resources: [input.yaml]\n", []diskFile{{"input.yaml", "kind: ConfigMap\n", 0644}}, nil},
	{"kustomize-generator", "kustomize", "Kustomization", "configMapGenerator:\n - name: acme\n   files: [settings=input.txt]\n", []diskFile{{"input.txt", "value", 0644}}, nil},
	{"kustomize-patch", "kustomize", "kustomization.yml", "patches:\n - path: patch.yaml\n", []diskFile{{"patch.yaml", "kind: ConfigMap\n", 0644}}, nil},
	{"kustomize-missing", "kustomize", "kustomization.yaml", "resources: [missing.yaml]\n", nil, []string{"kustomize_input_not_materialized"}},
	{"kustomize-remote", "kustomize", "kustomization.yaml", "resources: [https://example.com/acme/resource.yaml]\n", nil, []string{"kustomize_remote_input"}},
	{"kustomize-outside", "kustomize", "kustomization.yaml", "resources: [../../outside.yaml]\n", nil, []string{"kustomize_input_outside_captured_source"}},
	{"kustomize-absolute", "kustomize", "kustomization.yaml", "resources: [/outside.yaml]\n", nil, []string{"kustomize_input_outside_captured_source"}},
	{"kustomize-git", "kustomize", "kustomization.yaml", "resources: [.git/HEAD]\n", []diskFile{{".git/HEAD", "metadata", 0644}}, []string{"kustomize_git_metadata_input"}},
	{"kustomize-external", "kustomize", "kustomization.yaml", "helmCharts: []\n", nil, []string{"kustomize_external_dependencies_uninspected"}},
	{"kustomize-unknown-field", "kustomize", "kustomization.yaml", "futureField: true\n", nil, []string{"kustomize_field_uninspected:futureField"}},
	{"kustomize-multiline", "kustomize", "kustomization.yaml", "resources:\n - |\n    kind: ConfigMap\n", nil, []string{"kustomize_multiline_input_uninspected"}},
	{"kustomize-malformed", "kustomize", "kustomization.yaml", "resources: [", nil, []string{"kustomize_dependencies_unparseable"}},
	{"pulumi-runtime", "pulumi", "Pulumi.yaml", "name: acme\nruntime: nodejs\n", nil, []string{"pulumi_runtime_dependencies_uninspected"}},
}

func TestCaptureClosureControls(t *testing.T) {
	for _, fixture := range closureFixtures {
		t.Run(fixture.name, func(t *testing.T) {
			fs, repo := afero.NewMemMapFs(), "/repo"
			files := append([]diskFile{{fixture.path, fixture.data, 0644}}, fixture.extra...)
			writeTree(t, fs, repo, "src", files)
			expected := sourceidentity.New(fixture.kind, "config")
			expected.Root = "src"
			source := rootedSource(t, fs, repo)
			actual, err := sourceidentity.Capture(source, expected)
			require.NoError(t, err)
			checkCapture(t, actual, expected, files, fixture.reasons)
			require.NoError(t, sourceidentity.Verify(source, actual))
		})
	}
}

func TestCaptureTerraformModuleCoverageProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		fs, repo := afero.NewMemMapFs(), "/repo"
		hidden := hegel.Draw(t, hegel.Booleans())
		depth := hegel.Draw(t, hegel.Integers(0, 2))
		module := strings.Repeat("nested/", depth) + "child"
		var reasons []string
		if hidden {
			module = strings.Repeat("nested/", depth) + ".git/child"
			reasons = []string{"terraform_module_outside_captured_source"}
		}
		files := []diskFile{
			{"main.tf", fmt.Sprintf(`module "child" { source = "./%s" }`, module), 0644},
			{module + "/main.tf", `output "x" { value = 1 }`, 0644},
		}
		writeTree(t, fs, repo, "src", files)
		expected := sourceidentity.New("terraform", "config")
		expected.Root = "src"
		source := rootedSource(t, fs, repo)
		prepared, err := sourceidentity.Capture(source, expected)
		require.NoError(t, err)
		checkCapture(t, prepared, expected, files, reasons)
		files[1].data = `output "x" { value = 2 }`
		writeDiskFile(t, fs, filepath.Join(repo, "src"), files[1])
		changed, err := sourceidentity.Capture(source, expected)
		require.NoError(t, err)
		checkCapture(t, changed, expected, files, reasons)
		require.Equal(t, !hidden, prepared.Checksum != changed.Checksum)
		if hidden {
			require.Empty(t, prepared.EligibleChecksum())
			require.Empty(t, changed.EligibleChecksum())
		} else {
			require.Error(t, sourceidentity.Verify(source, prepared))
		}
	})
}

func TestCaptureClosureProperties(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"terraform", "helm", "kustomize", "pulumi"} {
		t.Run(kind, func(t *testing.T) {
			var choices []closureFixture
			for _, fixture := range closureFixtures {
				if fixture.kind == kind {
					choices = append(choices, fixture)
				}
			}
			hegel.Test(t, func(t *hegel.T) {
				fs, repo := afero.NewMemMapFs(), "/repo"
				expected := sourceidentity.New(kind, "config")
				expected.Root = "src"
				var files []diskFile
				var reasons []string
				count := hegel.Draw(t, hegel.Integers(1, 3))
				for n := 0; n < count; n++ {
					fixture := hegel.Draw(t, hegel.SampledFrom(choices))
					prefix := fmt.Sprintf("part-%d/", n)
					files = append(files, diskFile{prefix + fixture.path, fixture.data, 0644})
					for _, extra := range fixture.extra {
						extra.path = prefix + extra.path
						files = append(files, extra)
					}
					reasons = append(reasons, fixture.reasons...)
				}
				if kind == "terraform" {
					locked := hegel.Draw(t, hegel.Booleans())
					if locked {
						files = append(files, diskFile{".terraform.lock.hcl", "# root lock", 0644})
					}
					reasons = slices.DeleteFunc(reasons, func(reason string) bool { return reason == "terraform_providers_not_locked" })
					if !locked && slices.Contains(reasons, "terraform_provider_closure_uninspected") {
						reasons = append(reasons, "terraform_providers_not_locked")
					}
				}
				if hegel.Draw(t, hegel.Booleans()) {
					files = append(files, diskFile{"pointer.txt", "version https://git-lfs.github.com/spec/v1\noid sha256:abc\nsize 1\n", 0644})
					reasons = append(reasons, "git_lfs_not_materialized")
				}
				if hegel.Draw(t, hegel.Booleans()) {
					writeDiskFile(t, fs, repo, diskFile{".gitmodules", "[submodule \"acme\"]\npath = child\n", 0644})
					reasons = append(reasons, "git_submodules_uninspected")
				}
				if hegel.Draw(t, hegel.Booleans()) {
					expected.Coverage = "unknown"
					expected.Reasons = make([]string, 1, 4)
					expected.Reasons[0] = "preexisting_unknown"
					reasons = append(reasons, "preexisting_unknown")
				}
				writeTree(t, fs, repo, "src", files)
				before := cloneIdentity(expected)
				source := rootedSource(t, fs, repo)
				actual, err := sourceidentity.Capture(source, expected)
				require.NoError(t, err)
				checkCapture(t, actual, expected, files, reasons)
				require.Equal(t, before, expected)
				require.NoError(t, sourceidentity.Verify(source, actual))
				repeated, err := sourceidentity.Capture(source, actual)
				require.NoError(t, err)
				checkCapture(t, repeated, expected, files, reasons)
				if len(actual.Reasons) > 0 {
					actual.Reasons[0] = "mutated"
				}
				actual.Files[0].Path = "mutated"
				require.Equal(t, before, expected)
				checkCapture(t, repeated, expected, files, reasons)
			})
		})
	}
}

func TestCaptureAliasingProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		fs, repo := afero.NewMemMapFs(), "/repo"
		files := []diskFile{{"main.tf", `module "x" { source = "example.com/acme/module" }`, 0644}}
		writeTree(t, fs, repo, "src", files)
		expected := drawIdentity(t)
		expected.Root, expected.Kind, expected.Coverage = "src", "terraform", "unknown"
		fileBacking := []sourceidentity.File{{Path: "stale.txt", SHA256: sum256("stale")}, {Path: "sentinel", SHA256: sum256("")}}
		reasonBacking := []string{"preexisting_unknown", "unused", "sentinel", "unused"}
		expected.Files, expected.Reasons = fileBacking[:1], reasonBacking[:1]
		expected.Checksum = sum256(identityPreimage(expected))
		before := cloneIdentity(expected)
		fileSnapshot, reasonSnapshot := slices.Clone(fileBacking), slices.Clone(reasonBacking)
		source := rootedSource(t, fs, repo)
		actual, err := sourceidentity.Capture(source, expected)
		require.NoError(t, err)
		reasons := []string{"preexisting_unknown", "terraform_module_outside_captured_source"}
		checkCapture(t, actual, expected, files, reasons)
		repeated, err := sourceidentity.Capture(source, expected)
		require.NoError(t, err)
		require.NoError(t, sourceidentity.Verify(source, actual))
		require.Equal(t, before, expected)
		require.Equal(t, fileSnapshot, fileBacking)
		require.Equal(t, reasonSnapshot, reasonBacking)
		actual.Files[0].Path, actual.Reasons[0] = "mutated", "mutated"
		require.Equal(t, before, expected)
		require.Equal(t, fileSnapshot, fileBacking)
		require.Equal(t, reasonSnapshot, reasonBacking)
		checkCapture(t, repeated, expected, files, reasons)
	})
}

func makeLinkTree(t require.TestingT, repo string, depth int, absolute bool, target string) {
	fs := afero.NewOsFs()
	writeTree(t, fs, repo, "src", nil)
	for _, name := range []string{"a.txt", "b.txt"} {
		writeDiskFile(t, fs, repo, diskFile{"targets/" + name, "equal bytes", 0755})
	}
	for n := 0; n < depth; n++ {
		next := fmt.Sprintf("link-%d", n+1)
		if n == depth-1 {
			next = "../targets/" + target
		}
		if absolute {
			next = filepath.Join(repo, "src", next)
		}
		require.NoError(t, os.Symlink(next, filepath.Join(repo, "src", fmt.Sprintf("link-%d", n))))
	}
}

func TestCaptureSafeLinkProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		fs := afero.NewOsFs()
		parent, cleanup := temporaryRepo(t)
		defer cleanup()
		repo, other := filepath.Join(parent, "a"), filepath.Join(parent, "b")
		depth := hegel.Draw(t, hegel.Integers(1, 4))
		absolute := hegel.Draw(t, hegel.Booleans())
		makeLinkTree(t, repo, depth, absolute, "a.txt")
		makeLinkTree(t, other, depth, absolute, "a.txt")
		expected := sourceidentity.New("terraform", "config")
		expected.Root = "src"
		source, otherSource := rootedSource(t, fs, repo), rootedSource(t, fs, other)
		prepared, err := sourceidentity.Capture(source, expected)
		require.NoError(t, err)
		require.NoError(t, prepared.Validate())
		require.Equal(t, "complete", prepared.Coverage)
		require.NotEmpty(t, prepared.EligibleChecksum())
		require.Len(t, prepared.Files, depth)
		for n, file := range prepared.Files {
			require.Equal(t, fmt.Sprintf("link-%d", n), file.Path)
			require.True(t, file.Executable)
			var links []string
			for step := n; step < depth; step++ {
				target := fmt.Sprintf("link-%d", step+1)
				if step == depth-1 {
					target = "../targets/a.txt"
				}
				if absolute {
					target = filepath.ToSlash(filepath.Join("src", target))
				}
				links = append(links, fmt.Sprintf(`{"Path":"src/link-%d","Target":%s,"Absolute":%t}`, step, quote(target), absolute))
			}
			linkHash := sum256(fmt.Sprintf(`{"Links":[%s],"Target":"targets/a.txt"}`, strings.Join(links, ",")))
			require.Equal(t, sum256(fmt.Sprintf(`{"Link":%s,"Content":%s}`, quote(linkHash), quote(sum256("equal bytes")))), file.SHA256)
		}
		relocated, err := sourceidentity.Capture(otherSource, expected)
		require.NoError(t, err)
		require.Equal(t, prepared, relocated)
		require.NoError(t, sourceidentity.Verify(otherSource, prepared))
		writeDiskFile(t, fs, repo, diskFile{"targets/b.txt", "unreferenced change", 0644})
		require.NoError(t, sourceidentity.Verify(source, prepared))
		writeDiskFile(t, fs, repo, diskFile{"targets/b.txt", "equal bytes", 0755})
		last := filepath.Join(repo, "src", fmt.Sprintf("link-%d", depth-1))
		require.NoError(t, os.Remove(last))
		target := "../targets/b.txt"
		if absolute {
			target = filepath.Join(repo, "targets/b.txt")
		}
		require.NoError(t, os.Symlink(target, last))
		retargeted, err := sourceidentity.Capture(source, expected)
		require.NoError(t, err)
		require.NotEqual(t, prepared.Checksum, retargeted.Checksum)
		require.Error(t, sourceidentity.Verify(source, prepared))
		writeDiskFile(t, fs, repo, diskFile{"targets/b.txt", "changed bytes", 0755})
		require.Error(t, sourceidentity.Verify(source, retargeted))
	})
}

func TestCaptureDirectoryLinkProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		fs := afero.NewOsFs()
		repo, cleanup := temporaryRepo(t)
		defer cleanup()
		rootLink := hegel.Draw(t, hegel.Booleans())
		writeTree(t, fs, repo, "src", nil)
		for _, target := range []string{"a", "b"} {
			writeDiskFile(t, fs, repo, diskFile{target + "/file.txt", "equal bytes", 0644})
		}
		name, target := filepath.Join(repo, "src", "linked"), "../a"
		expected := sourceidentity.New("terraform", "config")
		expected.Root = "src"
		if rootLink {
			name, target, expected.Root = filepath.Join(repo, "alias"), "a", "alias"
		}
		require.NoError(t, os.Symlink(target, name))
		source := rootedSource(t, fs, repo)
		prepared, err := sourceidentity.Capture(source, expected)
		require.NoError(t, err)
		require.NoError(t, prepared.Validate())
		require.Equal(t, "unknown", prepared.Coverage)
		require.Equal(t, []string{"source_directory_symlink_closure_uninspected"}, prepared.Reasons)
		require.Empty(t, prepared.EligibleChecksum())
		require.NoError(t, sourceidentity.Verify(source, prepared))
		require.NoError(t, os.Remove(name))
		if rootLink {
			target = "b"
		} else {
			target = "../b"
		}
		require.NoError(t, os.Symlink(target, name))
		retargeted, err := sourceidentity.Capture(source, expected)
		require.NoError(t, err)
		require.NotEqual(t, prepared.Checksum, retargeted.Checksum)
		require.Error(t, sourceidentity.Verify(source, prepared))
	})
}

func TestCaptureUnsafeLinkProperties(t *testing.T) {
	hegel.Test(t, func(t *hegel.T) {
		fs := afero.NewOsFs()
		parent, cleanup := temporaryRepo(t)
		defer cleanup()
		repo := filepath.Join(parent, "repo")
		writeTree(t, fs, repo, "src", nil)
		writeDiskFile(t, fs, parent, diskFile{"outside.txt", "outside", 0644})
		target := hegel.Draw(t, hegel.SampledFrom([]string{"../../outside.txt", filepath.Join(parent, "outside.txt"), "link", "peer", ".", "missing"}))
		depth := hegel.Draw(t, hegel.Integers(1, 4))
		for n := 0; n < depth; n++ {
			next := "link"
			if n == 0 {
				next = target
			}
			name := "link"
			if n > 0 {
				name = fmt.Sprintf("chain-%d", n)
			}
			require.NoError(t, os.Symlink(next, filepath.Join(repo, "src", name)))
		}
		if target == "peer" {
			require.NoError(t, os.Symlink("link", filepath.Join(repo, "src", "peer")))
		}
		expected := sourceidentity.New("terraform", "config")
		expected.Root = "src"
		before := cloneIdentity(expected)
		actual, err := sourceidentity.Capture(rootedSource(t, fs, repo), expected)
		require.Error(t, err)
		require.Nil(t, actual)
		require.Equal(t, before, expected)
	})
}

func TestCaptureEscapeAndReturn(t *testing.T) {
	fs := afero.NewOsFs()
	parent := t.TempDir()
	repo := filepath.Join(parent, "repo")
	writeTree(t, fs, repo, "src", []diskFile{{"file.txt", "inside", 0644}})
	require.NoError(t, os.Mkdir(filepath.Join(parent, "outside"), 0755))
	require.NoError(t, os.Symlink("../../outside/../repo/src/file.txt", filepath.Join(repo, "src", "link")))
	expected := sourceidentity.New("terraform", "config")
	expected.Root = "src"
	actual, err := sourceidentity.Capture(rootedSource(t, fs, repo), expected)
	require.Error(t, err)
	require.Nil(t, actual)
}

func TestCaptureHelmArchiveRoot(t *testing.T) {
	var archive bytes.Buffer
	gz := gzip.NewWriter(&archive)
	tw := tar.NewWriter(gz)
	metadata := "apiVersion: v2\nname: acme\nversion: 1.0.0\n"
	require.NoError(t, tw.WriteHeader(&tar.Header{Name: "acme/Chart.yaml", Mode: 0644, Size: int64(len(metadata))}))
	_, err := tw.Write([]byte(metadata))
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	require.NoError(t, gz.Close())
	data := archive.String()
	fs, repo := afero.NewOsFs(), t.TempDir()
	writeDiskFile(t, fs, repo, diskFile{"archives/a.tgz", data, 0644})
	chart, err := loader.Load(filepath.Join(repo, "archives/a.tgz"))
	require.NoError(t, err)
	require.Equal(t, "acme", chart.Name())
	source := rootedSource(t, fs, repo)
	expected := sourceidentity.New("helm", "config")
	expected.Root = "archives/a.tgz"
	actual, err := sourceidentity.Capture(source, expected)
	require.NoError(t, err)
	checkCapture(t, actual, expected, []diskFile{{".", data, 0644}}, nil)
	require.NoError(t, sourceidentity.Verify(source, actual))
	for _, kind := range []string{"terraform", "pulumi", "kustomize"} {
		t.Run(kind+"-requires-directory", func(t *testing.T) {
			expected := sourceidentity.New(kind, "config")
			expected.Root = "archives/a.tgz"
			actual, err := sourceidentity.Capture(source, expected)
			require.ErrorContains(t, err, "source root is not a directory")
			require.Nil(t, actual)
		})
	}
	for _, reason := range []string{"helm_dependencies_resolved_at_build", "runner_source_evidence_missing"} {
		t.Run(reason, func(t *testing.T) {
			unknown := cloneIdentity(expected)
			unknown.MarkUnknown(reason)
			captured, err := sourceidentity.Capture(source, unknown)
			require.NoError(t, err)
			checkCapture(t, captured, unknown, []diskFile{{".", data, 0644}}, []string{reason})
			require.Equal(t, actual.Checksum, captured.Checksum)
			require.Empty(t, captured.EligibleChecksum())
		})
	}
	require.NoError(t, fs.MkdirAll(filepath.Join(repo, "charts"), 0755))
	require.NoError(t, os.Symlink("a.tgz", filepath.Join(repo, "archives/current.tgz")))
	require.NoError(t, os.Symlink("../archives/current.tgz", filepath.Join(repo, "charts/input.tgz")))
	linked := sourceidentity.New("helm", "config")
	linked.Root = "charts/input.tgz"
	prepared, err := sourceidentity.Capture(source, linked)
	require.NoError(t, err)
	require.NoError(t, prepared.Validate())
	require.NotEmpty(t, prepared.EligibleChecksum())
	require.Len(t, prepared.Files, 1)
	require.Equal(t, ".", prepared.Files[0].Path)
	require.NotEqual(t, sum256(data), prepared.Files[0].SHA256)
	require.NoError(t, sourceidentity.Verify(source, prepared))
	require.NoError(t, os.Remove(filepath.Join(repo, "archives/current.tgz")))
	require.NoError(t, os.Symlink("a.tgz", filepath.Join(repo, "archives/alias.tgz")))
	require.NoError(t, os.Symlink("alias.tgz", filepath.Join(repo, "archives/current.tgz")))
	retargeted, err := sourceidentity.Capture(source, linked)
	require.NoError(t, err)
	require.NotEqual(t, prepared.Checksum, retargeted.Checksum)
	require.Error(t, sourceidentity.Verify(source, prepared))
	writeDiskFile(t, fs, repo, diskFile{"archives/a.tgz", data + "\x00", 0644})
	require.Error(t, sourceidentity.Verify(source, retargeted))
}

func TestCaptureBoundaries(t *testing.T) {
	fs := afero.NewOsFs()
	repo := t.TempDir()
	source := rootedSource(t, fs, repo)
	actual, err := sourceidentity.Capture(source, nil)
	require.Error(t, err)
	require.Nil(t, actual)
	for _, scope := range []string{"../outside", "missing", "file.txt"} {
		writeDiskFile(t, fs, repo, diskFile{"file.txt", "not a directory", 0644})
		expected := sourceidentity.New("terraform", "config")
		expected.Root = scope
		actual, err := sourceidentity.Capture(source, expected)
		require.Error(t, err)
		require.Nil(t, actual)
	}
	listener, err := net.Listen("unix", filepath.Join(repo, "socket"))
	require.NoError(t, err)
	defer listener.Close()
	actual, err = sourceidentity.Capture(source, sourceidentity.New("terraform", "config"))
	require.Error(t, err)
	require.Nil(t, actual)
	require.NoError(t, sourceidentity.Verify(nil, nil))
	uncaptured := sourceidentity.New("terraform", "config")
	uncaptured.Seal()
	require.NoError(t, sourceidentity.Verify(nil, uncaptured))
	uncaptured.Checksum = "invalid"
	require.Error(t, sourceidentity.Verify(nil, uncaptured))
}
