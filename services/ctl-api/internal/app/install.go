package app

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type Install struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"created_by"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index:idx_app_install_name,unique" json:"-"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`
	Org   Org    `json:"-" faker:"-"`

	Name  string `json:"name" gorm:"notnull;index:idx_app_install_name,unique"`
	App   App    `swaggerignore:"true" json:"app"`
	AppID string `json:"app_id" gorm:"notnull;index:idx_app_install_name,unique"`

	AppSandboxConfigID string           `json:"-" swaggerignore:"true"`
	AppSandboxConfig   AppSandboxConfig `json:"app_sandbox_config"`

	AppRunnerConfigID string          `json:"-" swaggerignore:"true"`
	AppRunnerConfig   AppRunnerConfig `json:"app_runner_config"`

	InstallComponents  []InstallComponent  `json:"install_components,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
	InstallSandboxRuns []InstallSandboxRun `json:"install_sandbox_runs,omitempty" gorm:"constraint:OnDelete:CASCADE;"`
	InstallInputs      []InstallInputs     `json:"install_inputs" gorm:"constraint:OnDelete:CASCADE;"`
	InstallEvents      []InstallEvent      `json:"install_events" gorm:"constraint:OnDelete:CASCADE;"`

	AWSAccount   *AWSAccount   `json:"aws_account" gorm:"constraint:OnDelete:CASCADE;"`
	AzureAccount *AzureAccount `json:"azure_account" gorm:"constraint:OnDelete:CASCADE;"`

	// generated view current view

	InstallNumber     int              `json:"install_number" gorm:"->;-:migration"`
	SandboxStatus     SandboxRunStatus `json:"sandbox_status" gorm:"->;-:migration" swaggertype:"string"`
	ComponentStatuses pgtype.Hstore    `json:"component_statuses" gorm:"type:hstore;->;-:migration" swaggertype:"object,string"`

	// after queries

	CurrentInstallInputs     *InstallInputs      `json:"-" gorm:"-"`
	CompositeComponentStatus InstallDeployStatus `json:"composite_component_status" gorm:"-" swaggertype:"string"`
	RunnerStatus             string              `json:"runner_status" gorm:"-" swaggertype:"string"`

	// TODO(jm): deprecate these fields once the terraform provider has been updated

	Status            string `json:"status" gorm:"-"`
	StatusDescription string `json:"status_description" gorm:"-"`
}

func (i *Install) UseView() bool {
	return true
}

func (i *Install) ViewVersion() string {
	return "v3"
}

func (i *Install) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = domains.NewInstallID()
	}

	i.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	i.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}

// We want to report the status of the sandbox, the runner, and the components,
// and then roll that up into a high-level status for the install overall.
func (i *Install) AfterQuery(tx *gorm.DB) error {
	// get the runner status
	i.RunnerStatus = "ok"
	if len(i.InstallInputs) > 0 {
		i.CurrentInstallInputs = &i.InstallInputs[0]
	}

	// get the sandbox status
	if i.SandboxStatus == SandboxRunStatusUnknown || i.SandboxStatus == SandboxRunStatusEmpty {
		i.SandboxStatus = SandboxRunStatusQueued
	}

	// get the composite status of all the components
	i.CompositeComponentStatus = compositeComponentStatus(i.ComponentStatuses)

	// determine the status of the install based on all three
	i.Status = string(installStatus(i.SandboxStatus, i.RunnerStatus, i.CompositeComponentStatus))

	return nil
}

// installStatus coalesces a single status for the install, from the individual statuses of the sandbox, runner, and components
// It doesn't look like we can determine, based on only on the statuses, whether the install is provisioning or deprovisioning.
// We may need to update the event loop deprovision workflow to set this status.
// For now, this is good enough for clients to know if the install is busy, and shouldn't be interacted with.
func installStatus(sandboxStatus SandboxRunStatus, runnerStatus string, componentStatus InstallDeployStatus) string {
	fmt.Println(sandboxStatus, runnerStatus, componentStatus)

	// if any status is "error", return "error"
	if sandboxStatus == "error" || runnerStatus == "error" || componentStatus == "error" {
		return "error"
	}

	// if all statuses are "active" (or "ok"), return "active"
	// if there are no components, then "noop" counts as "active"
	if sandboxStatus == "active" && runnerStatus == "ok" && (componentStatus == "active" || componentStatus == "noop") {
		return "active"
	}

	// otherwise, return "provisioning"
	return "provisioning"
}

// compositeComponentStatus coalesces a single status from the statuses of the app's components.
// This is based on the components defined in the app config, not the components present in the install.
// Components may be present in an install's history that have been removed from the app.
func compositeComponentStatus(componentStatuses pgtype.Hstore) InstallDeployStatus {
	// if there are no components, then there are no operations to wait for
	if len(componentStatuses) == 0 {
		return InstallDeployStatusNoop
	}

	// check status of each component
	activecount := 0
	for _, status := range componentStatuses {
		switch InstallDeployStatus(*status) {
		case InstallDeployStatusOK:
			activecount++
		case InstallDeployStatusError:
			// if any components have failed, composite status should be "error"
			// we can stop immediately
			return InstallDeployStatusError
		}
	}

	// if all components are active, composite status should be "active"
	if activecount == len(componentStatuses) {
		return InstallDeployStatusOK
	}

	// if any components have not yet succeeded or failed, composite status should be "pending"
	return InstallDeployStatusPending
}
