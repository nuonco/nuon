package config

import (
	"fmt"

	"github.com/invopop/jsonschema"
	"github.com/mitchellh/mapstructure"

	"github.com/powertoolsdev/mono/pkg/config/source"
)

type InstallerConfig struct {
	Source string `mapstructure:"source,omitempty"`

	Name        string   `mapstructure:"name,omitempty" toml:"name"`
	Description string   `mapstructure:"description,omitempty" toml:"description"`
	Slug        string   `mapstructure:"slug,omitempty" toml:"slug"`
	Apps        []string `mapstructure:"apps,omitempty" toml:"apps"`

	DocumentationURL string `mapstructure:"documentation_url,omitempty" toml:"documentation_url"`
	CommunityURL     string `mapstructure:"community_url,omitempty" toml:"community_url"`
	HomepageURL      string `mapstructure:"homepage_url,omitempty" toml:"homepage_url"`
	GithubURL        string `mapstructure:"github_url,omitempty" toml:"github_url"`
	LogoURL          string `mapstructure:"logo_url,omitempty" toml:"logo_url"`
	FaviconURL       string `mapstructure:"favicon_url,omitempty" toml:"favicon_url"`

	OgImageURL          string `mapstructure:"og_image_url" toml:"og_image_url"`
	DemoURL             string `mapstructure:"demo_url" toml:"demo_url"`
	PostInstallMarkdown string `mapstructure:"post_install_markdown" toml:"post_install_markdown"`
	CopyrightMarkdown   string `mapstructure:"copyright_markdown" toml:"copyright_markdown"`
	FooterMarkdown      string `mapstructure:"footer_markdown" toml:"footer_markdown"`
}

func (a InstallerConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("source").Short("external configuration source").
		Long("Path to an external file containing installer configuration (YAML, JSON, or TOML)").
		Field("name").Short("installer name").
		Long("Human-readable name for the installer").
		Example("My SaaS Installer").
		Field("description").Short("installer description").
		Long("Detailed description of what this installer does").
		Example("Complete installer for My SaaS application").
		Field("slug").Short("URL-safe slug").
		Long("URL-friendly identifier for the installer").
		Example("my-saas-installer").
		Field("apps").Short("list of app names").
		Long("Array of app names to include in this installer").
		Example("api").
		Example("web-ui").
		Field("documentation_url").Short("documentation URL").
		Long("Link to the application documentation").
		Example("https://docs.example.com").
		Field("community_url").Short("community URL").
		Long("Link to community resources or forum").
		Example("https://community.example.com").
		Field("homepage_url").Short("homepage URL").
		Long("Link to the application homepage").
		Example("https://example.com").
		Field("github_url").Short("GitHub repository URL").
		Long("Link to the GitHub repository").
		Example("https://github.com/example/repo").
		Field("logo_url").Short("logo URL").
		Long("URL to the application logo image").
		Example("https://example.com/logo.png").
		Field("favicon_url").Short("favicon URL").
		Long("URL to the favicon image").
		Example("https://example.com/favicon.ico").
		Field("og_image_url").Short("OpenGraph image URL").
		Long("URL to the image displayed when sharing the installer on social media").
		Example("https://example.com/og-image.png").
		Field("demo_url").Short("demo URL").
		Long("Link to a live demo of the application").
		Example("https://demo.example.com").
		Field("post_install_markdown").Short("post-install markdown").
		Long("Markdown content displayed to users after successful installation").
		Field("copyright_markdown").Short("copyright markdown").
		Long("Markdown content for copyright information").
		Field("footer_markdown").Short("footer markdown").
		Long("Markdown content displayed in the installer footer")
}

func (a *InstallerConfig) Validate() error {
	if a == nil {
		return nil
	}

	return nil
}

func (a *InstallerConfig) parse() error {
	if a.Source == "" {
		return nil
	}

	obj, err := source.LoadSource(a.Source)
	if err != nil {
		return ErrConfig{
			Description: fmt.Sprintf("unable to load source %s", a.Source),
			Err:         err,
		}
	}

	if err := mapstructure.Decode(obj, &a); err != nil {
		return err
	}
	return nil
}
