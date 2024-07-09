package app

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type InstallerMetadata struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"created_by"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index"`

	InstallerID string `json:"installer_id" gorm:"notnull"`

	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`
	Org   Org    `json:"-" faker:"-"`

	Name        string `json:"name" gorm:"notnull"`
	Description string `json:"description" gorm:"notnull"`

	PostInstallMarkdown generics.NullString `json:"post_install_markdown" swaggertype:"string"`
	FooterMarkdown      generics.NullString `json:"footer_markdown" swaggertype:"string"`
	CopyrightMarkdown   generics.NullString `json:"copyright_markdown" swaggertype:"string"`
	DemoURL             generics.NullString `json:"demo_url" swaggertype:"string"`
	OgImageUrl          generics.NullString `json:"og_image_url"`

	DocumentationURL string `json:"documentation_url" gorm:"notnull"`
	LogoURL          string `json:"logo_url" gorm:"notnull"`
	GithubURL        string `json:"github_url" gorm:"notnull"`
	CommunityURL     string `json:"community_url" gorm:"notnull"`
	HomepageURL      string `json:"homepage_url" gorm:"notnull"`
	FaviconURL       string `json:"favicon_url"`

	FormattedDemoURL string `json:"formatted_demo_url" gorm:"-"`
}

func (a *InstallerMetadata) AfterQuery(tx *gorm.DB) error {
	a.FormattedDemoURL = a.DemoURL.String
	if !strings.HasPrefix(a.DemoURL.String, "https://www.youtube.com") {
		return nil
	}
	if strings.HasPrefix(a.DemoURL.String, "https://www.youtube.com/embed") {
		return nil
	}

	u, err := url.Parse(a.DemoURL.String)
	if err != nil {
		return nil
	}

	params, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return nil
	}

	ytID := params.Get("v")
	if ytID == "" {
		return nil
	}

	a.FormattedDemoURL = fmt.Sprintf("https://www.youtube.com/embed/%s", ytID)
	return nil
}

func (a *InstallerMetadata) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAppID()
	}

	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	return nil
}
