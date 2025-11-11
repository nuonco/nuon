package config

import (
	"github.com/invopop/jsonschema"
)

type MetadataConfig struct {
	// Config file version
	Version string `mapstructure:"version" jsonschema:"required"`

	// Description for your app, which is rendered in the installers
	Description string `mapstructure:"description,omitempty"`
	// Display name for the app, rendered in the installer
	DisplayName string `mapstructure:"display_name,omitempty"`
	// Slack webhook url to receive notifications
	SlackWebhookURL string `mapstructure:"slack_webhook_url"`
	// Readme for the app
	Readme string `mapstructure:"readme,omitempty"`
}

func (m MetadataConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("version").Short("config file version").Required().
		Long("Version of the configuration file format").
		Example("1.0.0").
		Example("2.0.0").
		Field("description").Short("app description").
		Long("Detailed description of the application, displayed in the installer UI").
		Example("A powerful SaaS platform for managing deployments").
		Field("display_name").Short("app display name").
		Long("Human-readable name for the application, shown in the installer").
		Example("My SaaS App").
		Example("Enterprise Platform").
		Field("slack_webhook_url").Short("Slack webhook URL").
		Long("Slack webhook URL to receive deployment notifications and updates").
		Example("https://hooks.slack.com/services/YOUR/WEBHOOK/URL").
		Field("readme").Short("README content").
		Long("Markdown content displayed as README documentation for the application")
}
