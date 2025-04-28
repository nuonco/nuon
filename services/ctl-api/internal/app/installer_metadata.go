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
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index" temporaljson:"deleted_at,omitzero,omitempty"`

	InstallerID string `json:"installer_id" gorm:"notnull" temporaljson:"installer_id,omitzero,omitempty"`

	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	Name        string `json:"name" gorm:"notnull" temporaljson:"name,omitzero,omitempty"`
	Description string `json:"description" gorm:"notnull" temporaljson:"description,omitzero,omitempty"`

	PostInstallMarkdown generics.NullString `json:"post_install_markdown" swaggertype:"string" temporaljson:"post_install_markdown,omitzero,omitempty"`
	FooterMarkdown      generics.NullString `json:"footer_markdown" swaggertype:"string" temporaljson:"footer_markdown,omitzero,omitempty"`
	CopyrightMarkdown   generics.NullString `json:"copyright_markdown" swaggertype:"string" temporaljson:"copyright_markdown,omitzero,omitempty"`
	DemoURL             generics.NullString `json:"demo_url" swaggertype:"string" temporaljson:"demo_url,omitzero,omitempty"`
	OgImageURL          generics.NullString `json:"og_image_url" swaggertype:"string" temporaljson:"og_image_url,omitzero,omitempty"`

	DocumentationURL string `json:"documentation_url" gorm:"notnull" temporaljson:"documentation_url,omitzero,omitempty"`
	LogoURL          string `json:"logo_url" gorm:"notnull" temporaljson:"logo_url,omitzero,omitempty"`
	GithubURL        string `json:"github_url" gorm:"notnull" temporaljson:"github_url,omitzero,omitempty"`
	CommunityURL     string `json:"community_url" gorm:"notnull" temporaljson:"community_url,omitzero,omitempty"`
	HomepageURL      string `json:"homepage_url" gorm:"notnull" temporaljson:"homepage_url,omitzero,omitempty"`
	FaviconURL       string `json:"favicon_url" temporaljson:"favicon_url,omitzero,omitempty"`

	FormattedDemoURL string `json:"formatted_demo_url" gorm:"-" temporaljson:"formatted_demo_url,omitzero,omitempty"`
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
