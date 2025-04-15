package config

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
