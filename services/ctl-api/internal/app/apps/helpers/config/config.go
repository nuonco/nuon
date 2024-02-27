package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

const (
	currentVersion string = "v1"
)

type AppConfig struct {
	Version string `mapstructure:"version"`

	Inputs    *AppInputConfig     `mapstructure:"inputs,omitempty"`
	Sandbox   *AppSandboxConfig   `mapstructure:"sandbox"`
	Runner    *AppRunnerConfig    `mapstructure:"runner"`
	Installer *AppInstallerConfig `mapstructure:"installer,omitempty"`

	Components []*Component `mapstructure:"components" validate:"gte=1"`
}

func (a *AppConfig) Validate(v *validator.Validate) error {
	if a.Version != currentVersion {
		return fmt.Errorf("version must be v1")
	}

	if err := v.Struct(a); err != nil {
		return err
	}

	return nil
}
