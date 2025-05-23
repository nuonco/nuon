package parse

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"github.com/spf13/afero"

	"github.com/powertoolsdev/mono/pkg/config"
	"github.com/powertoolsdev/mono/pkg/config/parse/dir"
	"github.com/powertoolsdev/mono/pkg/config/parse/get"
)

const (
	defaultFieldGetTimeout time.Duration = time.Second * 5
)

func parseDirName(dirname string) (string, error) {
	if _, err := os.Stat(dirname); err != nil {
		if os.IsNotExist(err) {
			return "", config.ErrConfig{
				Description: dirname + " does not exist",
				Err:         err,
			}
		}

		return "", err
	}

	absPath, err := filepath.Abs(dirname)
	if err != nil {
		return "", errors.Wrap(err, "unable to get absolute path")
	}

	return absPath, nil
}

func ParseDir(ctx context.Context, parseCfg ParseConfig) (*config.AppConfig, error) {
	fp, err := parseDirName(parseCfg.Dirname)
	if err != nil {
		return nil, err
	}

	fs := afero.NewOsFs()
	cfgFS := afero.NewBasePathFs(fs, fp)

	// parse the directory
	var obj ConfigDir
	if err := dir.Parse(ctx, cfgFS, &obj, &dir.ParseOptions{
		Root:     fp,
		Ext:      ".toml",
		ParserFn: func(rc io.ReadCloser, s string, a any) error { return parseTomlFile(rc, s, a, parseCfg.FileProcessor) },
	}); err != nil {
		return nil, errors.Wrap(err, "unable to parse directory")
	}

	// NOTE(jm): this will go away once we deprecate the legacy config, and we can just have a pipeline of
	// `config.AppConfig` parsers.
	appCfg, err := obj.toAppConfig()
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert to app config")
	}

	// parse all get functions
	if err := get.Parse(ctx, appCfg, &get.Options{
		FieldTimeout: defaultFieldGetTimeout,
		RootDir:      fp,
	}); err != nil {
		return nil, ParseErr{
			Description: "unable to get fields",
			Err:         err,
		}
	}

	err = appCfg.Parse()
	if err != nil {
		return nil, ParseErr{
			Description: "error parsing config",
			Err:         err,
		}
	}

	checksums, err := checksumTOMLFilesByName(cfgFS)
	if err != nil {
		return nil, errors.Wrap(err, "unable to checksum toml files")
	}

	for _, cmp := range appCfg.Components {
		if checksum, ok := checksums[cmp.Name]; ok {
			cmp.Checksum = checksum
		}
	}

	return appCfg, nil
}

func checksumTOMLFilesByName(cfgFS afero.Fs) (map[string]string, error) {
	checksums := make(map[string]string)

	// Read the components directory
	files, err := afero.ReadDir(cfgFS, "components")
	if err != nil {
		return nil, fmt.Errorf("failed to read components directory: %w", err)
	}

	for _, file := range files {
		// Skip directories and non-TOML files
		if file.IsDir() || !strings.HasSuffix(strings.ToLower(file.Name()), ".toml") {
			continue
		}

		filePath := filepath.Join("components", file.Name())

		// Read file content
		content, err := afero.ReadFile(cfgFS, filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", filePath, err)
		}

		// Parse TOML to get the name
		var config config.Component
		if err := toml.Unmarshal(content, &config); err != nil {
			return nil, fmt.Errorf("failed to parse TOML in %s: %w", filePath, err)
		}

		// Skip files without a name field
		if config.Name == "" {
			fmt.Printf("Warning: %s has no 'name' field, skipping\n", filePath)
			continue
		}

		// Calculate SHA256 checksum
		hash := sha256.Sum256(content)
		checksums[config.Name] = fmt.Sprintf("%x", hash)
	}

	return checksums, nil
}
