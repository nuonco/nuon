package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type Org struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"notnull"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index:idx_org_name,unique" json:"-"`

	Name              string `gorm:"index:idx_org_name,unique;notnull" json:"name"`
	Status            string `json:"status" gorm:"notnull"`
	StatusDescription string `json:"status_description" gorm:"notnull"`

	// These fields are used to control the behaviour of the org
	// NOTE: these are starting as nullable, so we can update stage/prod before resetting locally.
	SandboxMode bool `json:"sandbox_mode" gorm:"notnull"`
	CustomCert  bool `json:"custom_cert" gorm:"notnull"`

	Apps           []App            `faker:"-" swaggerignore:"true" json:"apps,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
	VCSConnections []VCSConnection  `json:"vcs_connections,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
	UserOrgs       []UserOrg        `json:"users,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
	HealthChecks   []OrgHealthCheck `json:"health_checks,omitempty" gorm:"constraint:OnDelete:CASCADE;"`

	// Filled in at read time
	LatestHealthCheck OrgHealthCheck `json:"latest_health_check" gorm:"-"`
}

func (o *Org) AfterQuery(tx *gorm.DB) error {
	if len(o.HealthChecks) < 1 {
		return nil
	}

	o.LatestHealthCheck = o.HealthChecks[0]
	return nil
}

func (o *Org) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = domains.NewOrgID()
	}

	o.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	return nil
}
