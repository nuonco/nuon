package config

type AppInstallerConfig struct {
	Name        string `mapstructure:"name" toml:"name"`
	Description string `mapstructure:"description" toml:"description"`
	Slug        string `mapstructure:"slug" toml:"slug"`

	DocumentationURL string `mapstructure:"documentation_url" toml:"documentation_url"`
	CommunityURL     string `mapstructure:"community_url" toml:"community_url"`
	HomepageURL      string `mapstructure:"homepage_url" toml:"homepage_url"`
	GithubURL        string `mapstructure:"github_url" toml:"github_url"`
	LogoURL          string `mapstructure:"logo_url" toml:"logo_url"`
	DemoURL          string `mapstructure:"demo_url" toml:"demo_url"`

	PostInstallMarkdown string `mapstructure:"post_install_markdown" toml:"post_install_markdown"`
}

func (t *AppInstallerConfig) ToResourceType() string {
	return "nuon_app_installer"
}

func (t *AppInstallerConfig) ToResource() (map[string]interface{}, error) {
	resource, err := toMapStructure(t)
	if err != nil {
		return nil, err
	}
	if resource == nil {
		return nil, nil
	}

	resource["app_id"] = "${var.app_id}"

	return nestWithName("installer", resource), nil
}
