package app

import (
	"time"

	"gorm.io/gorm"

	customermanaged "github.com/nuonco/nuon/pkg/runner/customer_managed"
	"github.com/nuonco/nuon/pkg/shortid/domains"
)

const (
	InstallRegistrationSourceManual         = "manual"
	InstallRegistrationSourceConnectedAgent = "connected_agent"

	InstallConnectivityDisconnected = "disconnected"
	InstallConnectivityConnected    = "connected"

	InstallReleaseSelectionCustomer = "customer"
	InstallReleaseSelectionVendor   = "vendor_proposed"

	InstallAuthorityCustomer = "customer"
	InstallAuthorityNuon     = "nuon"

	InstallTelemetryManual      = "manual"
	InstallTelemetryOperational = "operational"
	InstallTelemetryLive        = "live"

	InstallDeploymentMethodDisconnectedLocal = "disconnected_local"
	InstallDeploymentMethodConnectedLocal    = "connected_local"
	InstallDeploymentMethodNuonManaged       = "nuon_managed"

	InstallDeploymentStatusSucceeded = "succeeded"
	InstallDeploymentStatusFailed    = "failed"
	InstallDeploymentStatusCancelled = "cancelled"
	InstallDeploymentStatusSkipped   = "skipped"
)

type InstallManagementPolicyVersion struct {
	ID           string     `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID  string     `json:"created_by_id" gorm:"not null;default:null"`
	CreatedAt    time.Time  `json:"created_at" gorm:"notnull"`
	EffectiveAt  time.Time  `json:"effective_at" gorm:"notnull"`
	SupersededAt *time.Time `json:"superseded_at,omitempty"`

	OrgID     string  `json:"-" gorm:"notnull;index"`
	InstallID string  `json:"install_id" gorm:"notnull;uniqueIndex:idx_install_management_policy_version"`
	Install   Install `json:"-" gorm:"constraint:OnDelete:RESTRICT;"`
	Version   int     `json:"version" gorm:"notnull;uniqueIndex:idx_install_management_policy_version"`

	Connectivity      string `json:"connectivity" gorm:"notnull"`
	ReleaseSelection  string `json:"release_selection" gorm:"notnull"`
	CommandAuthority  string `json:"command_authority" gorm:"notnull"`
	ApprovalAuthority string `json:"approval_authority" gorm:"notnull"`
	Telemetry         string `json:"telemetry" gorm:"notnull"`
}

func (p *InstallManagementPolicyVersion) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = domains.NewInstallManagementPolicyVersionID()
	}
	if p.CreatedByID == "" {
		p.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if p.OrgID == "" {
		p.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}

type InstallRegistration struct {
	ID          string    `gorm:"primary_key" json:"id"`
	CreatedByID string    `json:"created_by_id" gorm:"not null;default:null"`
	CreatedAt   time.Time `json:"created_at" gorm:"notnull"`

	OrgID     string         `json:"-" gorm:"notnull;index"`
	InstallID string         `json:"install_id" gorm:"notnull;index"`
	Install   Install        `json:"-" gorm:"constraint:OnDelete:RESTRICT;"`
	ReleaseID string         `json:"release_id" gorm:"notnull;index"`
	Release   AppRelease     `json:"-" gorm:"constraint:OnDelete:RESTRICT;"`
	PackageID string         `json:"package_id" gorm:"notnull;index"`
	Package   ReleasePackage `json:"-" gorm:"constraint:OnDelete:RESTRICT;"`

	Source            string                                   `json:"source" gorm:"notnull"`
	Registration      customermanaged.InstallationRegistration `json:"registration" gorm:"type:jsonb;serializer:json;<-:create"`
	IntegrityStatus   string                                   `json:"integrity_status" gorm:"notnull"`
	AssociationStatus string                                   `json:"association_status" gorm:"notnull"`
	ImportedAt        time.Time                                `json:"imported_at" gorm:"notnull"`
}

func (r *InstallRegistration) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = r.Registration.RegistrationID
	}
	if r.CreatedByID == "" {
		r.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if r.OrgID == "" {
		r.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}

type InstallReleaseDeployment struct {
	ID                        string                         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedAt                 time.Time                      `json:"created_at" gorm:"notnull"`
	OrgID                     string                         `json:"-" gorm:"notnull;index"`
	InstallID                 string                         `json:"install_id" gorm:"notnull;index;uniqueIndex:idx_install_release_operation"`
	Install                   Install                        `json:"-" gorm:"constraint:OnDelete:RESTRICT;"`
	ReleaseID                 string                         `json:"release_id" gorm:"notnull;index"`
	Release                   AppRelease                     `json:"-" gorm:"constraint:OnDelete:RESTRICT;"`
	PackageID                 *string                        `json:"package_id,omitempty" gorm:"index"`
	Package                   *ReleasePackage                `json:"-" gorm:"constraint:OnDelete:RESTRICT;"`
	PreviousReleaseID         string                         `json:"previous_release_id,omitempty"`
	InstallAppConfigVersionID *string                        `json:"install_app_config_version_id,omitempty" gorm:"index"`
	InstallAppConfigVersion   *InstallAppConfigVersion       `json:"-" gorm:"constraint:OnDelete:RESTRICT;"`
	PolicyVersionID           string                         `json:"policy_version_id" gorm:"notnull;index"`
	PolicyVersion             InstallManagementPolicyVersion `json:"-" gorm:"constraint:OnDelete:RESTRICT;"`

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
