package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type NotificationsConfig struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"notnull"`
	CreatedBy   UserToken             `json:"created_by" gorm:"references:Subject"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `json:"-"`

	OrgID string `json:"org_id" gorm:"notnull;defaultnull"`

	OwnerID   string `json:"owner_id" gorm:"notnull;defaultnull;"`
	OwnerType string `json:"owner_type" gorm:"notnull;defaultnull;"`

	// slack settings
	EnableSlackNotifications bool   `json:"-" temporaljson:"enable_slack_notifications"`
	SlackWebhookURL          string `json:"slack_webhook_url"`
	InternalSlackWebhookURL  string `json:"-" temporaljson:"internal_slack_webhook_url"`

	// email settings
	EnableEmailNotifications bool `json:"-" temporaljson:"enable_email_notifications"`

	// generated via after query
	SlackWebhookURLs []string `gorm:"-" json:"-" temporal_json:"slack_webhook_urls"`
}

func (a *NotificationsConfig) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAppID()
	}

	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	return nil
}

func (a *NotificationsConfig) AfterQuery(tx *gorm.DB) error {
	a.SlackWebhookURLs = []string{
		a.InternalSlackWebhookURL,
	}
	if a.SlackWebhookURL != "" {
		a.SlackWebhookURLs = append(a.SlackWebhookURLs, a.SlackWebhookURL)
	}

	return nil
}
