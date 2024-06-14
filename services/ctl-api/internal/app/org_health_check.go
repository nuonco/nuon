package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type OrgHealthCheckStatus string

const (
	OrgHealthCheckStatusOK         OrgHealthCheckStatus = "ok"
	OrgHealthCheckStatusError      OrgHealthCheckStatus = "error"
	OrgHealthCheckStatusInProgress OrgHealthCheckStatus = "in-progress"

	// for denoting when a job is going on, and should poll.
	OrgHealthCheckStatusProvisioning   OrgHealthCheckStatus = "provisioning"
	OrgHealthCheckStatusDeprovisioning OrgHealthCheckStatus = "deprovisioning"
)

type OrgHealthCheck struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"created_by"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `json:"-"`

	Status            OrgHealthCheckStatus `json:"status" gorm:"notnull"`
	StatusDescription string               `json:"status_description" gorm:"notnull"`

	OrgID string `gorm:"notnull"`
}

func (o *OrgHealthCheck) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = domains.NewOrgID()
	}

	return nil
}
