package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type VCSWebhookSubscription struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `swaggerignore:"true" json:"-" temporaljson:"org,omitzero,omitempty"`

	VCSConnectionID string `json:"vcs_connection_id" gorm:"not null" temporaljson:"vcs_connection_id,omitzero,omitempty"`
	GithubInstallID string `json:"github_install_id" gorm:"uniqueIndex:idx_vcs_webhook_sub_install_id;not null;default:''" temporaljson:"github_install_id,omitzero,omitempty"`
	GithubHookID    int64  `json:"github_hook_id,omitempty" gorm:"default:null" temporaljson:"github_hook_id,omitzero,omitempty"`
	WebhookURL      string `json:"webhook_url,omitempty" gorm:"not null;default:''" temporaljson:"webhook_url,omitzero,omitempty"`
	WebhookSecret   string `json:"-" gorm:"not null;default:''" temporaljson:"webhook_secret,omitzero,omitempty"`

	Status *CompositeStatus `json:"status,omitempty" gorm:"column:status;type:jsonb;default:null" temporaljson:"status,omitzero,omitempty"`
}

func (v *VCSWebhookSubscription) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name:    indexes.Name(db, &VCSWebhookSubscription{}, "vcs_connection_id"),
			Columns: []string{"vcs_connection_id"},
		},
		{
			Name:    indexes.Name(db, &VCSWebhookSubscription{}, "github_install_id"),
			Columns: []string{"github_install_id"},
		},
	}
}

func (v *VCSWebhookSubscription) BeforeCreate(tx *gorm.DB) error {
	v.ID = domains.NewVCSWebhookSubscriptionID()
	if v.OrgID == "" {
		v.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	if v.CreatedByID == "" {
		v.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	return nil
}
