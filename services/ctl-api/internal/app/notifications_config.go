package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type NotificationsConfig struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id" gorm:"notnull" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id" gorm:"notnull;defaultnull" temporaljson:"org_id,omitzero,omitempty"`

	OwnerID   string `json:"owner_id" gorm:"notnull;defaultnull;" temporaljson:"owner_id,omitzero,omitempty"`
	OwnerType string `json:"owner_type" gorm:"notnull;defaultnull;" temporaljson:"owner_type,omitzero,omitempty"`

	// slack settings
	EnableSlackNotifications bool   `json:"-" temporaljson:"enable_slack_notifications,omitzero,omitempty"`
	SlackWebhookURL          string `json:"slack_webhook_url" temporaljson:"slack_webhook_url,omitzero,omitempty"`
	InternalSlackWebhookURL  string `json:"-" temporaljson:"internal_slack_webhook_url,omitzero,omitempty"`

	// email settings
	EnableEmailNotifications bool `json:"-" temporaljson:"enable_email_notifications,omitzero,omitempty"`

	// generated via after query
	SlackWebhookURLs []string `gorm:"-" json:"-" temporaljson:"slack_webhook_ur_ls,omitzero,omitempty"`
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
