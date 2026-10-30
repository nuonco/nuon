package sourceidentity

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"gopkg.in/yaml.v3"

	"github.com/nuonco/nuon/pkg/sourcefs"
)

var commitPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

func IsCommit(ref string) bool { return commitPattern.MatchString(ref) }

func Capture(source *sourcefs.FS, expected *Identity) (*Identity, error) {
	if expected == nil {
		return nil, fmt.Errorf("source identity is required")
	}
	out := *expected
	out.Files = nil
	out.Reasons = append([]string(nil), expected.Reasons...)
	root := "."
	path, rootLinks, err := source.ResolveWithLinks(filepath.FromSlash(expected.Root))
	if err != nil {
		return nil, err
	}
	addFile := func(file File) error {
		out.Files = append(out.Files, file)
		if len(out.Files) > 20000 {
			return fmt.Errorf("source manifest exceeds file limit")
		}
		return nil
	}
	directories := make(map[string]bool)
	var walk func(string, string, []sourcefs.Link) error
	walk = func(name, rel string, links []sourcefs.Link) error {
		info, err := source.Lstat(name)
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			name, links, err = source.ResolveWithLinks(name)
			if err != nil {
				return err
			}
			info, err = source.Lstat(name)
			if err != nil {
				return err
			}
		}
		sum := ""
		if len(links) > 0 {
			target, err := filepath.Rel(root, name)
			if err != nil {
				return err
			}
			sum = Hash(struct {
				Links  []sourcefs.Link
				Target string
			}{links, filepath.ToSlash(target)})
		}
		if info.IsDir() {
			if directories[name] {
				return fmt.Errorf("source directory symlink loop at %s", name)
			}
			directories[name] = true
			defer delete(directories, name)
			if len(links) > 0 {
				out.MarkUnknown("source_directory_symlink_closure_uninspected")
				if err := addFile(File{Path: filepath.ToSlash(rel), SHA256: sum}); err != nil {
					return err
				}
			}
			entries, err := source.ReadDir(name)
			if err != nil {
				return err
			}
			for _, entry := range entries {
				if entry.Name() == ".git" {
					continue
				}
				if err := walk(filepath.Join(name, entry.Name()), filepath.Join(rel, entry.Name()), nil); err != nil {
					return err
				}
			}
			return nil
		}
		if rel == "." && out.Kind != "helm" {
			return fmt.Errorf("source root is not a directory")
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("source file %s is not regular", name)
		}
		data, err := source.ReadFile(name)
		if err != nil {
			return err
		}
		if len(links) > 0 {
			// Equal materialized bytes must not hide a changed link chain or target.
			sum = Hash(struct{ Link, Content string }{sum, sha256Bytes(data)})
		} else {
			sum = sha256Bytes(data)
		}
		if err := addFile(File{Path: filepath.ToSlash(rel), SHA256: sum, Executable: info.Mode().Perm()&0111 != 0}); err != nil {
			return err
		}
		if strings.HasPrefix(string(data), "version https://git-lfs.github.com/spec/") {
			out.MarkUnknown("git_lfs_not_materialized")
		}
		name = filepath.Join(path, rel)
		switch out.Kind {
		case "terraform":
			if strings.HasSuffix(name, ".tf") {
				inspectTerraform(source, &out, name, data, path)
			}
			if strings.HasSuffix(name, ".tf.json") {
				out.MarkUnknown("terraform_json_dependencies_uninspected")
			}
		case "helm":
			if filepath.Base(name) == "Chart.yaml" || filepath.Base(name) == "requirements.yaml" {
				inspectChart(&out, data)
			}
		case "kustomize":
			if strings.EqualFold(filepath.Base(name), "kustomization.yaml") || strings.EqualFold(filepath.Base(name), "kustomization.yml") || filepath.Base(name) == "Kustomization" {
				inspectKustomize(source, &out, data, filepath.Dir(name), path)
			}
		}
		return nil
	}
	if err := walk(path, ".", rootLinks); err != nil {
		return nil, err
	}
	if _, err := source.Lstat(".gitmodules"); err == nil {
		out.MarkUnknown("git_submodules_uninspected")
	}
	if out.Kind == "pulumi" {
		out.MarkUnknown("pulumi_runtime_dependencies_uninspected")
	}
	out.Captured = true
	out.Seal()
	return &out, nil
}

func Verify(source *sourcefs.FS, expected *Identity) error {
	if expected == nil {
		return nil
	}
	if err := expected.Validate(); err != nil {
		return err
	}
	if !expected.Captured {
		return nil
	}
	actual, err := Capture(source, expected)
	if err != nil {
		return err
	}
	if !Matches(expected, actual) {
		return fmt.Errorf("captured source differs from prepared build inputs")
	}
	return nil
}

func within(root, path string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

func inspectTerraform(source *sourcefs.FS, i *Identity, name string, data []byte, root string) {
	file, diags := hclsyntax.ParseConfig(data, name, hcl.InitialPos)
	if diags.HasErrors() {
		i.MarkUnknown("terraform_dependencies_unparseable")
		return
	}
	body := file.Body.(*hclsyntax.Body)
	for _, block := range body.Blocks {
		if block.Type == "module" {
			attr := block.Body.Attributes["source"]
			if attr == nil {
				i.MarkUnknown("terraform_module_source_unknown")
				continue
			}
			value, diag := attr.Expr.Value(nil)
			if diag.HasErrors() || !value.IsKnown() || value.IsNull() || !value.Type().Equals(cty.String) {
				i.MarkUnknown("terraform_module_source_unknown")
				continue
			}
			source := value.AsString()
			if (!strings.HasPrefix(source, "./") && !strings.HasPrefix(source, "../")) || !within(root, filepath.Join(filepath.Dir(name), source)) || strings.Contains("/"+filepath.ToSlash(source)+"/", "/.git/") {
				i.MarkUnknown("terraform_module_outside_captured_source")
			}
		}
		if block.Type == "provider" || block.Type == "terraform" || block.Type == "resource" || block.Type == "data" {
			if _, _, err := source.ResolveWithin(root, filepath.Join(root, ".terraform.lock.hcl")); err != nil {
				i.MarkUnknown("terraform_providers_not_locked")
			}
			i.MarkUnknown("terraform_provider_closure_uninspected")
		}
		if block.Type == "data" {
			i.MarkUnknown("terraform_runtime_inputs_uninspected")
		}
	}
	_ = hclsyntax.VisitAll(body, func(node hclsyntax.Node) hcl.Diagnostics {
		if call, ok := node.(*hclsyntax.FunctionCallExpr); ok {
			switch call.Name {
			case "file", "filebase64", "fileexists", "fileset", "templatefile", "filesha256", "filesha512", "filemd5", "filebase64sha256", "filebase64sha512", "pathexpand", "abspath":
				i.MarkUnknown("terraform_filesystem_expression_uninspected")
			}
		}
		return nil
	})
}

func inspectChart(i *Identity, data []byte) {
	var chart struct {
		Dependencies []any `yaml:"dependencies"`
	}
	if err := yaml.Unmarshal(data, &chart); err != nil {
		i.MarkUnknown("helm_dependencies_unparseable")
		return
	}
	if len(chart.Dependencies) > 0 {
		i.MarkUnknown("helm_dependencies_resolved_at_build")
	}
}

func inspectKustomize(source *sourcefs.FS, i *Identity, data []byte, directory, root string) {
	var obj map[string]any
	if err := yaml.Unmarshal(data, &obj); err != nil {
		i.MarkUnknown("kustomize_dependencies_unparseable")
		return
	}
	for key, value := range obj {
		switch key {
		case "apiVersion", "kind", "namespace", "namePrefix", "nameSuffix", "labels", "commonLabels", "commonAnnotations", "images", "replicas", "generatorOptions", "buildMetadata", "sortOptions":
		case "helmCharts", "helmGlobals", "generators", "transformers", "validators":
			i.MarkUnknown("kustomize_external_dependencies_uninspected")
		case "resources", "bases", "components", "patchesStrategicMerge", "patches", "patchesJson6902", "configMapGenerator", "secretGenerator", "configurations", "crds", "replacements", "openapi":
			inspectPaths(source, i, value, directory, root)
		default:
			i.MarkUnknown("kustomize_field_uninspected:" + key)
		}
	}
}

func inspectPaths(source *sourcefs.FS, i *Identity, value any, directory, root string) {
	switch v := value.(type) {
	case string:
		if strings.Contains(v, "://") || strings.Contains(v, "git::") || strings.Contains(v, "github.com/") || strings.Contains(v, "?ref=") {
			i.MarkUnknown("kustomize_remote_input")
			return
		}
		if strings.Contains(v, "\n") {
			i.MarkUnknown("kustomize_multiline_input_uninspected")
			return
		}
		_, candidate, found := strings.Cut(v, "=")
		if !found {
			candidate = v
		}
		if filepath.IsAbs(candidate) || !within(root, filepath.Join(directory, candidate)) {
			i.MarkUnknown("kustomize_input_outside_captured_source")
			return
		}
		if strings.Contains("/"+filepath.ToSlash(candidate)+"/", "/.git/") {
			i.MarkUnknown("kustomize_git_metadata_input")
			return
		}
		if _, _, err := source.ResolveWithin(root, filepath.Join(directory, candidate)); err != nil {
			i.MarkUnknown("kustomize_input_not_materialized")
		}
	case []any:
		for _, child := range v {
			inspectPaths(source, i, child, directory, root)
		}
	case map[string]any:
		for key, child := range v {
			switch key {
			case "path", "files", "envs", "env", "resources":
				inspectPaths(source, i, child, directory, root)
			}
		}
	}
}
