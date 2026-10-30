package kubernetes_manifest

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/yaml"

	"github.com/nuonco/nuon/pkg/sourcefs"
)

var kustomizeRemoteReference = regexp.MustCompile(`(?i)^(https?:|ssh://|file://|git::|github[.]com[:/])|^[a-zA-Z][a-zA-Z0-9-]*@[^/\\]*[:/\\]`)

// Kustomize's absolute paths are labels in this filesystem, not host paths.
type kustomizeFS struct {
	source         *sourcefs.FS
	kustomizations map[string]string
	pluginConfigs  map[string]bool
	err            error
}

var _ filesys.FileSystem = (*kustomizeFS)(nil)

func (f *kustomizeFS) Open(name string) (filesys.File, error) {
	return f.source.Open(name)
}

func (f *kustomizeFS) ReadFile(name string) ([]byte, error) {
	data, err := f.source.ReadFile(name)
	if err != nil {
		return nil, err
	}
	directory, known := f.kustomizations[name]
	if known || slices.Contains(konfig.RecognizedKustomizationFileNames(), filepath.Base(name)) {
		if !known {
			directory = filepath.Dir(name)
		}
		// Validate the exact bytes returned to Kustomize, not a separate preflight
		// read that could be replaced before Kustomize interprets it.
		if err := f.localKustomization(directory, data); err != nil {
			f.err = fmt.Errorf("kustomization %q: %w", name, err)
			return nil, f.err
		}
	}
	if err := localKustomizePlugins(data); err != nil {
		f.err = fmt.Errorf("plugin configuration %q: %w", name, err)
		return nil, f.err
	}
	return data, nil
}

func (f *kustomizeFS) CleanedAbs(name string) (filesys.ConfirmedDir, string, error) {
	input, err := f.source.Relative(name)
	if err != nil {
		return "", "", err
	}
	rel, err := f.source.Resolve(name)
	if err != nil {
		return "", "", err
	}
	info, err := f.source.Lstat(rel)
	if err != nil {
		return "", "", err
	}
	abs := filepath.Join(f.source.Root(), rel)
	if info.IsDir() {
		if f.pluginConfigs[input] {
			// Directory kustomizations can synthesize plugin paths after byte
			// validation, bypassing FileSystem via the builtin HTTP loader.
			f.err = fmt.Errorf("directory plugin configuration %q requires confined evaluation and is not supported in rooted builds; use a configuration file", name)
			return "", "", f.err
		}
		return filesys.ConfirmedDir(abs), "", nil
	}
	if slices.Contains(konfig.RecognizedKustomizationFileNames(), filepath.Base(name)) {
		if f.kustomizations == nil {
			f.kustomizations = make(map[string]string)
		}
		f.kustomizations[abs] = filepath.Dir(name)
	}
	return filesys.ConfirmedDir(filepath.Dir(abs)), filepath.Base(abs), nil
}

func (f *kustomizeFS) IsDir(name string) bool {
	_, base, err := f.CleanedAbs(name)
	return err == nil && base == ""
}

func (f *kustomizeFS) Exists(name string) bool {
	_, err := f.source.Resolve(name)
	return err == nil
}

func (f *kustomizeFS) ReadDir(name string) ([]string, error) {
	entries, err := f.source.ReadDir(name)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(entries))
	for i, entry := range entries {
		names[i] = entry.Name()
	}
	return names, nil
}

func (f *kustomizeFS) Glob(pattern string) ([]string, error) {
	matches, err := f.source.Glob(pattern)
	if err != nil {
		return nil, err
	}
	for i := range matches {
		matches[i] = filepath.Join(f.source.Root(), matches[i])
	}
	if !filesys.IsHiddenFilePath(pattern) {
		matches = filesys.RemoveHiddenFiles(matches)
	}
	return matches, nil
}

func (f *kustomizeFS) Walk(name string, fn filepath.WalkFunc) error {
	return f.source.Walk(name, func(rel string, info os.FileInfo, err error) error {
		return fn(filepath.Join(f.source.Root(), rel), info, err)
	})
}

func (f *kustomizeFS) Create(name string) (filesys.File, error) {
	return nil, &fs.PathError{Op: "create", Path: name, Err: fs.ErrPermission}
}

func (f *kustomizeFS) Mkdir(name string) error {
	return &fs.PathError{Op: "mkdir", Path: name, Err: fs.ErrPermission}
}

func (f *kustomizeFS) MkdirAll(name string) error { return f.Mkdir(name) }

func (f *kustomizeFS) RemoveAll(name string) error {
	return &fs.PathError{Op: "remove", Path: name, Err: fs.ErrPermission}
}

func (f *kustomizeFS) WriteFile(name string, _ []byte) error {
	return &fs.PathError{Op: "write", Path: name, Err: fs.ErrPermission}
}

func (f *kustomizeFS) localKustomization(directory string, data []byte) error {
	var k types.Kustomization
	if err := k.Unmarshal(data); err != nil {
		return err
	}
	if len(k.HelmCharts) > 0 || len(k.HelmChartInflationGenerator) > 0 {
		return fmt.Errorf("Helm inflation requires confined subprocess execution and is not supported in rooted builds")
	}
	paths := append([]string{k.OpenAPI["path"]}, k.Resources...)
	paths = append(paths, k.Bases...)
	paths = append(paths, k.Components...)
	paths = append(paths, k.Crds...)
	paths = append(paths, k.Configurations...)
	for _, patch := range append(k.Patches, k.PatchesJson6902...) {
		paths = append(paths, patch.Path)
	}
	for _, replacement := range k.Replacements {
		paths = append(paths, replacement.Path)
	}
	for _, generator := range k.ConfigMapGenerator {
		if err := localKustomizeGenerator(generator.KvPairSources); err != nil {
			return err
		}
	}
	for _, generator := range k.SecretGenerator {
		if err := localKustomizeGenerator(generator.KvPairSources); err != nil {
			return err
		}
	}
	for _, name := range paths {
		if err := localKustomizeReference(name); err != nil {
			return err
		}
	}
	resources := resmap.NewFactory(provider.NewDepProvider().GetResourceFactory())
	for _, patch := range k.PatchesStrategicMerge {
		if _, err := resources.NewResMapFromBytes([]byte(patch)); err == nil {
			continue
		}
		if err := localKustomizeReference(string(patch)); err != nil {
			return err
		}
	}
	for _, paths := range [][]string{k.Generators, k.Transformers, k.Validators} {
		for _, name := range paths {
			// Kustomize accepts either an inline resource or a file/base here.
			if _, err := resources.NewResMapFromBytes([]byte(name)); err == nil {
				if err := localKustomizePlugins([]byte(name)); err != nil {
					return err
				}
				continue
			}
			if err := localKustomizeReference(name); err != nil {
				return err
			}
			if !filepath.IsAbs(name) {
				name = filepath.Join(directory, name)
			}
			rel, err := f.source.Relative(name)
			if err != nil {
				return err
			}
			if f.pluginConfigs == nil {
				f.pluginConfigs = make(map[string]bool)
			}
			f.pluginConfigs[rel] = true
		}
	}
	return nil
}

func localKustomizeGenerator(sources types.KvPairSources) error {
	paths := append([]string{sources.EnvSource}, sources.EnvSources...)
	for _, source := range sources.FileSources {
		_, name, found := strings.Cut(source, "=")
		if !found {
			name = source
		}
		paths = append(paths, name)
	}
	for _, name := range paths {
		if err := localKustomizeReference(name); err != nil {
			return err
		}
	}
	return nil
}

// FileLoader fetches HTTP inputs before consulting FileSystem, including paths
// inside builtin plugin configurations. Keep those reads on the source too.
func localKustomizePlugins(data []byte) error {
	resources := resmap.NewFactory(provider.NewDepProvider().GetResourceFactory())
	rm, err := resources.NewResMapFromBytes(data)
	if err != nil {
		return nil
	}
	for _, resource := range rm.Resources() {
		if resource.GetApiVersion() != "builtin" {
			continue
		}
		config, err := resource.AsYAML()
		if err != nil {
			return err
		}
		var paths []string
		switch resource.GetKind() {
		case "ConfigMapGenerator", "SecretGenerator":
			var sources types.KvPairSources
			if err := yaml.Unmarshal(config, &sources); err != nil {
				return err
			}
			if err := localKustomizeGenerator(sources); err != nil {
				return err
			}
		case "PatchTransformer", "PatchJson6902Transformer", "PatchStrategicMergeTransformer", "ReplacementTransformer", "ValueAddTransformer":
			var inputs struct {
				Path           string                   `json:"path"`
				Paths          []string                 `json:"paths"`
				Replacements   []types.ReplacementField `json:"replacements"`
				TargetFilePath string                   `json:"targetFilePath"`
			}
			if err := yaml.Unmarshal(config, &inputs); err != nil {
				return err
			}
			paths = append(inputs.Paths, inputs.Path, inputs.TargetFilePath)
			for _, replacement := range inputs.Replacements {
				paths = append(paths, replacement.Path)
			}
		}
		for _, name := range paths {
			if _, err := resources.NewResMapFromBytes([]byte(name)); err == nil {
				continue
			}
			if err := localKustomizeReference(name); err != nil {
				return err
			}
		}
	}
	return nil
}

func localKustomizeReference(name string) error {
	if filepath.IsAbs(name) || strings.HasPrefix(name, "./") || strings.HasPrefix(name, "../") {
		return nil
	}
	if kustomizeRemoteReference.MatchString(name) {
		return fmt.Errorf("remote Kustomize reference %q requires confined fetching and is not supported in rooted builds; vendor it into the repository", name)
	}
	return nil
}
