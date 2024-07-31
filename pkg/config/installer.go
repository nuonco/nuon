package config

import (
	"fmt"

	"github.com/mitchellh/mapstructure"

	"github.com/powertoolsdev/mono/pkg/config/source"
)

type InstallerConfig struct {
	Source string `mapstructure:"source,omitempty"`

	Name        string   `mapstructure:"name,omitempty" toml:"name"`
	Description string   `mapstructure:"description,omitempty" toml:"description"`
	Slug        string   `mapstructure:"slug,omitempty" toml:"slug"`
	Apps        []string `mapstructure:"apps,omitempty" toml:"apps"`
	AppIDs      []string `mapstructure:"app_ids,omitempty" toml:"-"`

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

func (a *InstallerConfig) Validate() error {
	if a == nil {
		return nil
	}

	if len(a.AppIDs) > 0 {
		return ErrConfig{
			Description: "please use `apps` instead",
		}
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
