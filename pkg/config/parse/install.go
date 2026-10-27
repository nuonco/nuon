package parse

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/mitchellh/mapstructure"
	"github.com/pelletier/go-toml/v2"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/parse/get"
)

func ParseInstallConfig(r io.Reader, v *validator.Validate) (*config.Install, error) {
	decoder := toml.NewDecoder(r)

	obj := make(map[string]interface{})
	if err := decoder.Decode(&obj); err != nil {
		return nil, fmt.Errorf("error decoding TOML: %w", err)
	}

	var cfg config.Install
	decCfg := config.DecoderConfig()
	decCfg.Result = &cfg
	dec, err := mapstructure.NewDecoder(decCfg)
	if err != nil {
		return nil, err
	}

	if err := dec.Decode(obj); err != nil {
		return nil, fmt.Errorf("error decoding config: %w", err)
	}

	if err := cfg.Parse(); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("error validating config: %w", err)
	}

	return &cfg, nil
}

// LoadInstallUIDefinition fetches an install override's ui_definition into
// UIDefinitionContents. An unset path is a no-op. The config syncer uploads
// whatever this leaves in the field.
//
// Relative paths resolve from installDir, and local reads are confined to
// rootDir, the checked-out repository.
func LoadInstallUIDefinition(ctx context.Context, cfg *config.Install, rootDir, installDir string) error {
	if cfg == nil || cfg.StackOverrides == nil || strings.TrimSpace(cfg.StackOverrides.UIDefinition) == "" {
		return nil
	}
	cfg.StackOverrides.UIDefinitionContents = cfg.StackOverrides.UIDefinition

	workDir, err := filepath.Rel(rootDir, installDir)
	if err != nil || workDir == ".." || strings.HasPrefix(workDir, ".."+string(filepath.Separator)) {
		return fmt.Errorf("install directory %q is outside %q", installDir, rootDir)
	}
	if err := get.Parse(ctx, cfg.StackOverrides, &get.Options{
		FieldTimeout: defaultFieldGetTimeout,
		RootDir:      rootDir,
		WorkDir:      workDir,
		Restricted:   true,
	}); err != nil {
		return fmt.Errorf("ui_definition: %w", err)
	}
	return nil
}
