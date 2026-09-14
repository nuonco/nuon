package app

import (
	"time"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid/domains"
)

const (
	InstallConnectivityConnected = "connected"

	InstallReleaseSelectionCustomer = "customer"
	InstallReleaseSelectionVendor   = "vendor_proposed"

	InstallAuthorityCustomer = "customer"
	InstallAuthorityVendor   = "vendor"

	InstallTelemetryOperational = "operational"
	InstallTelemetryLive        = "live"
)

type InstallOperatingModel struct {
	ID          string    `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string    `json:"created_by_id" gorm:"not null;default:null"`
	CreatedAt   time.Time `json:"created_at" gorm:"notnull"`

	OrgID     string  `json:"-" gorm:"notnull;index"`
	InstallID string  `json:"install_id" gorm:"notnull;uniqueIndex"`
	Install   Install `json:"-" gorm:"constraint:OnDelete:RESTRICT;"`

	Connectivity      string `json:"connectivity" gorm:"notnull"`
	ReleaseSelection  string `json:"release_selection" gorm:"notnull"`
	ApprovalAuthority string `json:"approval_authority" gorm:"notnull"`
	Telemetry         string `json:"telemetry" gorm:"notnull"`
}

func (m *InstallOperatingModel) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = domains.NewInstallOperatingModelID()
	}
	if m.CreatedByID == "" {
		m.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if m.OrgID == "" {
		m.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}
