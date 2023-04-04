package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const (
	configFilename string = "service.yml"
)

func (l *loader) Load() (*Config, error) {
	fp := filepath.Join(fmt.Sprintf("%s/services/%s/%s", l.RootDir, l.Service, configFilename))
	cfg, err := l.load(fp)
	if err != nil {
		return nil, fmt.Errorf("unable to load filepath %s: %w", fp, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("unable to validate config: %w", err)
	}

	return cfg, nil
}

func (l *loader) load(fp string) (*Config, error) {
	byts, err := os.ReadFile(fp)
	if err != nil {
		return nil, fmt.Errorf("unable to read file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(byts, &cfg); err != nil {
		return nil, fmt.Errorf("unable to parse config: %w", err)
	}

	return &cfg, nil
}
