package parse

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/pelletier/go-toml/v2"

	"github.com/nuonco/nuon/pkg/config"
)

// AppBranchConfigFile pairs a parsed branch config with the file it came from,
// so callers can point at the offending file when two files declare the same
// branch name.
type AppBranchConfigFile struct {
	Path   string
	Config *config.AppBranchConfig
}

// ParseAppBranchConfig reads a single standalone app branch config from a TOML
// document. Unlike the app config directory parser, the document is the branch
// itself rather than a `[branch]` section of a larger config.
func ParseAppBranchConfig(r io.Reader) (*config.AppBranchConfig, error) {
	obj := make(map[string]interface{})
	if err := toml.NewDecoder(r).Decode(&obj); err != nil {
		return nil, fmt.Errorf("error decoding TOML: %w", err)
	}

	var cfg config.AppBranchConfig
	decCfg := config.DecoderConfig()
	decCfg.Result = &cfg
	dec, err := mapstructure.NewDecoder(decCfg)
	if err != nil {
		return nil, err
	}

	if err := dec.Decode(obj); err != nil {
		return nil, fmt.Errorf("error decoding config: %w", err)
	}

	if cfg.Name == "" {
		return nil, fmt.Errorf("branch config is missing a name")
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// ParseAppBranchConfigFile parses one branch config file.
func ParseAppBranchConfigFile(path string) (*config.AppBranchConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unable to read %s: %w", path, err)
	}
	defer f.Close()

	cfg, err := ParseAppBranchConfig(f)
	if err != nil {
		return nil, ParseErr{
			Filename:    path,
			Description: err.Error(),
			Err:         err,
		}
	}
	return cfg, nil
}

// ParseAppBranchConfigDir recursively parses every *.toml file under dir. Files
// are visited in sorted path order so the resulting plan is deterministic, and
// two files declaring the same branch name are rejected.
func ParseAppBranchConfigDir(dir string) ([]AppBranchConfigFile, error) {
	var paths []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.EqualFold(filepath.Ext(d.Name()), ".toml") {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unable to read branch configs from %s: %w", dir, err)
	}
	sort.Strings(paths)

	configs := make([]AppBranchConfigFile, 0, len(paths))
	seen := make(map[string]string, len(paths))
	for _, path := range paths {
		cfg, err := ParseAppBranchConfigFile(path)
		if err != nil {
			return nil, err
		}
		if prev, ok := seen[cfg.Name]; ok {
			return nil, ParseErr{
				Filename:    path,
				Description: fmt.Sprintf("branch %q is already declared in %s", cfg.Name, prev),
			}
		}
		seen[cfg.Name] = path
		configs = append(configs, AppBranchConfigFile{Path: path, Config: cfg})
	}

	return configs, nil
}

// LoadAppBranchConfigs loads a single TOML file or a directory of them.
func LoadAppBranchConfigs(path string) ([]AppBranchConfigFile, bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, fmt.Errorf("path %q does not exist", path)
		}
		return nil, false, fmt.Errorf("unable to access %s: %w", path, err)
	}

	if info.IsDir() {
		files, err := ParseAppBranchConfigDir(path)
		return files, true, err
	}

	cfg, err := ParseAppBranchConfigFile(path)
	if err != nil {
		return nil, false, err
	}
	return []AppBranchConfigFile{{Path: path, Config: cfg}}, false, nil
}
