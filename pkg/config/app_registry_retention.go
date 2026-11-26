package config

import (
	"fmt"
	"slices"

	"github.com/invopop/jsonschema"
)

const (
	RegistryRetentionModeOff     = "off"
	RegistryRetentionModeDryRun  = "dry_run"
	RegistryRetentionModeEnforce = "enforce"

	DefaultRegistryRetentionMode            = RegistryRetentionModeOff
	DefaultRegistryRetentionRetainCount     = 3
	DefaultRegistryRetentionGracePeriodDays = 30
)

var registryRetentionModes = []string{
	RegistryRetentionModeOff,
	RegistryRetentionModeDryRun,
	RegistryRetentionModeEnforce,
}

func ValidateRegistryRetentionMode(mode *string) error {
	if mode == nil || slices.Contains(registryRetentionModes, *mode) {
		return nil
	}
	return ErrConfig{
		Description: fmt.Sprintf("registry_retention mode %q is invalid, must be one of: off, dry_run, enforce", *mode),
	}
}

func ValidateRegistryRetentionRetainCount(count *int) error {
	if count == nil || *count >= 1 {
		return nil
	}
	return ErrConfig{
		Description: fmt.Sprintf("registry_retention retain_count %d is invalid, must be at least 1", *count),
	}
}

func ValidateRegistryRetentionGracePeriodDays(days *int) error {
	if days == nil || *days >= 0 {
		return nil
	}
	return ErrConfig{
		Description: fmt.Sprintf("registry_retention grace_period_days %d is invalid, must be 0 or more", *days),
	}
}

type RegistryRetention struct {
	Mode            *string `mapstructure:"mode,omitempty" toml:"mode,omitempty" json:"mode,omitempty"`
	RetainCount     *int    `mapstructure:"retain_count,omitempty" toml:"retain_count,omitempty" json:"retain_count,omitempty"`
	GracePeriodDays *int    `mapstructure:"grace_period_days,omitempty" toml:"grace_period_days,omitempty" json:"grace_period_days,omitempty"`
}

func (a RegistryRetention) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("mode").Short("registry retention mode").
		Long("Controls pruning of old images from each install's registry. off never prunes, dry_run reports what would be pruned without deleting, enforce deletes. Installs can override the mode. Leave unset to use the platform default.").
		Example("off").
		Example("dry_run").
		Example("enforce").
		Field("retain_count").Short("images to retain").
		Long("Number of most recent images to keep per component in each install's registry. Must be at least 1. Components can override this value. Leave unset to use the platform default.").
		Example(3).
		Example(10).
		Field("grace_period_days").Short("days to keep recent images").
		Long("Images synced or deployed within this many days are never pruned, regardless of retain_count. Must be 0 or more. Leave unset to use the platform default of 30.").
		Example(30).
		Example(7)
}

func (a *RegistryRetention) parse() error {
	return nil
}

func (a *RegistryRetention) Validate() error {
	if a == nil {
		return nil
	}
	if err := ValidateRegistryRetentionMode(a.Mode); err != nil {
		return err
	}
	if err := ValidateRegistryRetentionRetainCount(a.RetainCount); err != nil {
		return err
	}
	return ValidateRegistryRetentionGracePeriodDays(a.GracePeriodDays)
}

type ComponentRegistryRetention struct {
	RetainCount *int `mapstructure:"retain_count,omitempty" toml:"retain_count,omitempty" json:"retain_count,omitempty" nuonhash:"omitempty"`
}

func (a ComponentRegistryRetention) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("retain_count").Short("images to retain").
		Long("Number of most recent images to keep for this component in each install's registry. Must be at least 1. Overrides the app level retain_count. Leave unset to inherit it.").
		Example(10)
}

func (a *ComponentRegistryRetention) Validate() error {
	if a == nil {
		return nil
	}
	return ValidateRegistryRetentionRetainCount(a.RetainCount)
}

type InstallRegistryRetention struct {
	Mode *string `mapstructure:"mode,omitempty" toml:"mode,omitempty" json:"mode,omitempty"`
}

func (a InstallRegistryRetention) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("mode").Short("registry retention mode override").
		Long("Overrides the app level registry retention mode for this install. off never prunes, dry_run reports what would be pruned without deleting, enforce deletes. Omit to inherit the app setting.").
		Example("off").
		Example("dry_run").
		Example("enforce")
}

func (a *InstallRegistryRetention) Validate() error {
	if a == nil {
		return nil
	}
	return ValidateRegistryRetentionMode(a.Mode)
}
