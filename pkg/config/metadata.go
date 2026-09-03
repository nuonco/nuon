package config

import (
	"github.com/invopop/jsonschema"
)

type MetadataConfig struct {
	// Config file version
	Version string `mapstructure:"version" toml:"version" jsonschema:"required"`

	// Description for your app, which is rendered in the installers
	Description string `mapstructure:"description,omitempty" toml:"description,omitempty"`
	// Display name for the app, rendered in the installer
	DisplayName string `mapstructure:"display_name,omitempty" toml:"display_name,omitempty"`
	// Slack webhook url to receive notifications
	SlackWebhookURL string `mapstructure:"slack_webhook_url" toml:"slack_webhook_url"`
	// Readme for the app
	Readme string `mapstructure:"readme,omitempty" toml:"readme,omitempty"`
	// Color codes for label keys, keyed by label key name
	LabelColors map[string]string `mapstructure:"label_colors,omitempty" toml:"label_colors,omitempty"`
	// Labels applied to every install of the app; editable only via app config
	DefaultLabels   map[string]string      `mapstructure:"default_labels,omitempty" toml:"default_labels,omitempty"`
	CustomerManaged *CustomerManagedConfig `mapstructure:"customer_managed,omitempty" toml:"customer_managed,omitempty"`
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
		Long("Markdown content displayed as README documentation for the application").
		Example("./README.md").
		Field("label_colors").Short("label key color codes").
		Long("Map of label key names to hex color codes for customizing label display in the dashboard").
		Example(`{"env": "#FF5733", "region": "#33FF57"}`).
		Field("default_labels").Short("default labels for all installs").
		Long("Labels applied to every install of the app. Values may use the interpolation syntax ({{ .nuon.* }}). These labels cannot be edited or removed on individual installs — only via the app config").
		Example(`{"tier": "prod", "region": "{{ .nuon.cloud_account.aws.region }}"}`).
		Field("customer_managed").Short("customer-managed package runtime").
		Long("Vendor-controlled runtime artifacts included in customer-managed release packages")
}
