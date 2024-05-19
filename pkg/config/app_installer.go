package config

type InstallerConfig struct {
	Name        string   `mapstructure:"name,omitempty" toml:"name"`
	Description string   `mapstructure:"description,omitempty" toml:"description"`
	Slug        string   `mapstructure:"slug,omitempty" toml:"slug"`
	AppIDs      []string `mapstructure:"app_ids,omitempty" toml:"app_ids"`

	DocumentationURL string `mapstructure:"documentation_url,omitempty" toml:"documentation_url"`
	CommunityURL     string `mapstructure:"community_url,omitempty" toml:"community_url"`
	HomepageURL      string `mapstructure:"homepage_url,omitempty" toml:"homepage_url"`
	GithubURL        string `mapstructure:"github_url,omitempty" toml:"github_url"`
	LogoURL          string `mapstructure:"logo_url,omitempty" toml:"logo_url"`
	DemoURL          string `mapstructure:"demo_url,omitempty" toml:"demo_url"`
	FaviconURL       string `mapstructure:"favicon_url,omitempty" toml:"favicon_url"`

	PostInstallMarkdown string `mapstructure:"post_install_markdown,omitempty" toml:"post_install_markdown"`
	CopyrightMarkdown   string `mapstructure:"copyright_markdown,omitempty" toml:"copyright_markdown"`
	FooterMarkdown      string `mapstructure:"footer_markdown,omitempty" toml:"footer_markdown"`
}

func (a *InstallerConfig) ToResourceType() string {
	return "installer"
}

func (a *InstallerConfig) ToResource() (map[string]interface{}, error) {
	resource, err := toMapStructure(a)
	if err != nil {
		return nil, err
	}
	if resource == nil {
		return nil, nil
	}

	return nestWithName("installer", resource), nil
}

func (a *InstallerConfig) parse() error {
	return nil
}
