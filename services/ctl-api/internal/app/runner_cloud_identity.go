package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type RunnerCloudIdentity struct {
	ID          string  `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string  `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account `json:"-" temporaljson:"created_by,omitzero,omitempty"`

	CreatedAt time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"index:idx_runner_cloud_identity_runner,unique,priority:2" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" temporaljson:"org,omitzero,omitempty"`

	RunnerID string `json:"runner_id,omitzero" gorm:"not null;default:null;index:idx_runner_cloud_identity_runner,unique,priority:1" temporaljson:"runner_id,omitzero,omitempty"`

	Provider  CloudPlatform `json:"provider,omitzero" gorm:"not null;default:null" swaggertype:"string" temporaljson:"provider,omitzero,omitempty"`
	AccountID string        `json:"account_id,omitzero" gorm:"not null;default:null" temporaljson:"account_id,omitzero,omitempty"`
	Region    string        `json:"region,omitzero" gorm:"not null;default:null" temporaljson:"region,omitzero,omitempty"`

	RunnerRoleARN string `json:"runner_role_arn,omitzero" temporaljson:"runner_role_arn,omitzero,omitempty"`

	VendorRoleARN string `json:"vendor_role_arn,omitzero" temporaljson:"vendor_role_arn,omitzero,omitempty"`
	VendorRoleID  string `json:"vendor_role_id,omitzero" temporaljson:"vendor_role_id,omitzero,omitempty"`
}

func (r *RunnerCloudIdentity) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = domains.NewRunnerCloudIdentityID()
	}
	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if r.OrgID == "" {
		r.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}

func (r *RunnerCloudIdentity) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name:    indexes.Name(db, &RunnerCloudIdentity{}, "org_id"),
			Columns: []string{"org_id"},
		},
	}
}
