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

	InstallDeploymentMethodVendorManaged = "vendor_managed"

	InstallDeploymentStatusSucceeded = "succeeded"
	InstallDeploymentStatusFailed    = "failed"
	InstallDeploymentStatusCancelled = "cancelled"
	InstallDeploymentStatusSkipped   = "skipped"
)

func VendorManagedInstalls(db *gorm.DB) *gorm.DB {
	return db.Where(`id NOT IN (
		SELECT install_id
		FROM install_operating_models
		WHERE approval_authority <> ?
	)`, InstallAuthorityVendor)
}

func (i *Install) AppBranchUpdateEligible() bool {
	if i.OperatingModel == nil {
		return true
	}
	if i.OperatingModel.ApprovalAuthority == InstallAuthorityVendor {
		return true
	}
	return i.OperatingModel.Connectivity == InstallConnectivityConnected &&
		i.OperatingModel.ReleaseSelection == InstallReleaseSelectionVendor &&
		i.OperatingModel.ApprovalAuthority == InstallAuthorityCustomer
}

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

func (p *InstallOperatingModel) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = domains.NewInstallOperatingModelID()
	}
	if p.CreatedByID == "" {
		p.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if p.OrgID == "" {
		p.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}

type InstallReleaseDeployment struct {
	ID                        string                   `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedAt                 time.Time                `json:"created_at" gorm:"notnull"`
	OrgID                     string                   `json:"-" gorm:"notnull;index"`
	InstallID                 string                   `json:"install_id" gorm:"notnull;index;uniqueIndex:idx_install_release_operation"`
	Install                   Install                  `json:"-" gorm:"constraint:OnDelete:RESTRICT;"`
	ReleaseID                 string                   `json:"release_id" gorm:"notnull;index"`
	Release                   AppRelease               `json:"-" gorm:"constraint:OnDelete:RESTRICT;"`
	PackageID                 *string                  `json:"package_id,omitempty" gorm:"index"`
	Package                   *ReleasePackage          `json:"-" gorm:"constraint:OnDelete:RESTRICT;"`
	PreviousReleaseID         string                   `json:"previous_release_id,omitempty"`
	InstallAppConfigVersionID *string                  `json:"install_app_config_version_id,omitempty" gorm:"index"`
	InstallAppConfigVersion   *InstallAppConfigVersion `json:"-" gorm:"constraint:OnDelete:RESTRICT;"`
	OperatingModelID          string                   `json:"operating_model_id" gorm:"notnull;index"`
	OperatingModel            InstallOperatingModel    `json:"-" gorm:"constraint:OnDelete:RESTRICT;"`

	Method          string     `json:"method" gorm:"notnull"`
	Actor           string     `json:"actor" gorm:"notnull"`
	Executor        string     `json:"executor" gorm:"notnull"`
	OperationID     string     `json:"operation_id" gorm:"notnull;uniqueIndex:idx_install_release_operation"`
	PlanDigest      string     `json:"plan_digest,omitempty"`
	ResultDirective string     `json:"result_directive" gorm:"notnull"`
	Status          string     `json:"status" gorm:"notnull"`
	StartedAt       time.Time  `json:"started_at" gorm:"notnull"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"`
}

func (d *InstallReleaseDeployment) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = domains.NewInstallReleaseDeploymentID()
	}
	if d.OrgID == "" {
		d.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}
