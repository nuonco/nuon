package config

type AppInstallerConfig struct {
	Name        string `mapstructure:"name,omitempty" toml:"name"`
	Description string `mapstructure:"description,omitempty" toml:"description"`
	Slug        string `mapstructure:"slug,omitempty" toml:"slug"`

	DocumentationURL string `mapstructure:"documentation_url,omitempty" toml:"documentation_url"`
	CommunityURL     string `mapstructure:"community_url,omitempty" toml:"community_url"`
	HomepageURL      string `mapstructure:"homepage_url,omitempty" toml:"homepage_url"`
	GithubURL        string `mapstructure:"github_url,omitempty" toml:"github_url"`
	LogoURL          string `mapstructure:"logo_url,omitempty" toml:"logo_url"`
	DemoURL          string `mapstructure:"demo_url,omitempty" toml:"demo_url"`

	PostInstallMarkdown string `mapstructure:"post_install_markdown,omitempty" toml:"post_install_markdown"`
}

func (a *AppInstallerConfig) ToResourceType() string {
	return "nuon_app_installer"
}

func (a *AppInstallerConfig) ToResource() (map[string]interface{}, error) {
	resource, err := toMapStructure(a)
	if err != nil {
		return nil, err
	}
	if resource == nil {
		return nil, nil
	}

	resource["app_id"] = "${var.app_id}"

	return nestWithName("installer", resource), nil
}

func (a *AppInstallerConfig) parse() error {
	return nil
}
