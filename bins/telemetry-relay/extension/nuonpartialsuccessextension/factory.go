package nuonpartialsuccessextension

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
)

type Config struct{}

var componentType = component.MustNewType("nuonpartialsuccess")

func NewFactory() extension.Factory {
	return extension.NewFactory(
		componentType,
		createDefaultConfig,
		createExtension,
		component.StabilityLevelDevelopment,
	)
}

func createDefaultConfig() component.Config {
	return &Config{}
}

func createExtension(_ context.Context, settings extension.Settings, _ component.Config) (extension.Extension, error) {
	return &partialSuccessExtension{logger: settings.Logger}, nil
}
