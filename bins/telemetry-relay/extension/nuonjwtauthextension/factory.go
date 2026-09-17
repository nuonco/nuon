package nuonjwtauthextension

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
)

var componentType = component.MustNewType("nuonjwtauth")

func NewFactory() extension.Factory {
	return extension.NewFactory(
		componentType,
		createDefaultConfig,
		createExtension,
		component.StabilityLevelDevelopment,
	)
}

func createDefaultConfig() component.Config {
	return &Config{Audience: defaultAudience}
}

func createExtension(_ context.Context, settings extension.Settings, rawConfig component.Config) (extension.Extension, error) {
	cfg := rawConfig.(*Config)
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return newExtension(*cfg, settings.Logger), nil
}
