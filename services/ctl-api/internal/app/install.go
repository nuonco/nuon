package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/viewsql"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/bulk"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/links"
)

type Install struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index:idx_app_install_name,unique" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`
	Metadata    pgtype.Hstore         `json:"metadata,omitzero" gorm:"type:hstore" swaggertype:"object,string" temporaljson:"metadata,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	Name  string `json:"name,omitzero" gorm:"notnull;index:idx_app_install_name,unique" temporaljson:"name,omitzero,omitempty"`
	App   App    `swaggerignore:"true" json:"app,omitzero" temporaljson:"app,omitzero,omitempty"`
	AppID string `json:"app_id,omitzero" gorm:"notnull;index:idx_app_install_name,unique" temporaljson:"app_id,omitzero,omitempty"`

	AppConfigID string    `json:"app_config_id,omitzero" temporaljson:"app_config_id,omitzero,omitempty"`
	AppConfig   AppConfig `json:"-" temporaljson:"app_config,omitzero,omitempty"`

	AppSandboxConfigID string           `json:"-" swaggerignore:"true" temporaljson:"app_sandbox_config_id,omitzero,omitempty"`
	AppSandboxConfig   AppSandboxConfig `json:"app_sandbox_config,omitzero" temporaljson:"app_sandbox_config,omitzero,omitempty"`

	AppRunnerConfigID string          `json:"-" swaggerignore:"true" temporaljson:"app_runner_config_id,omitzero,omitempty"`
	AppRunnerConfig   AppRunnerConfig `json:"app_runner_config,omitzero" temporaljson:"app_runner_config,omitzero,omitempty"`

	InstallComponents       []InstallComponent        `json:"install_components,omitzero,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_components,omitzero,omitempty"`
	InstallActionWorkflows  []InstallActionWorkflow   `json:"install_action_workflows,omitzero,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_action_workflows,omitzero,omitempty"`
	InstallSandboxRuns      []InstallSandboxRun       `json:"install_sandbox_runs,omitzero,omitempty" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_sandbox_runs,omitzero,omitempty"`
	InstallInputs           []InstallInputs           `json:"install_inputs,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_inputs,omitzero,omitempty"`
	InstallEvents           []InstallEvent            `json:"install_events,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_events,omitzero,omitempty"`
	InstallIntermediateData []InstallIntermediateData `json:"-" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_intermediate_data,omitzero,omitempty"`
	InstallSandbox          InstallSandbox            `json:"sandbox" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_sandbox,omitzero,omitempty"`
	InstallConfig           *InstallConfig            `json:"install_config,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_config,omitzero,omitempty"`
	InstallStates           []InstallState            `json:"install_states,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_states,omitzero,omitempty"`

	InstallStack *InstallStack `json:"install_stack,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"install_stack,omitzero,omitempty"`
	AWSAccount   *AWSAccount   `json:"aws_account,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"aws_account,omitzero,omitempty"`
	AzureAccount *AzureAccount `json:"azure_account,omitzero" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"azure_account,omitzero,omitempty"`

	RunnerGroup RunnerGroup `json:"-" gorm:"polymorphic:Owner;constraint:OnDelete:CASCADE;" temporaljson:"runner_group,omitzero,omitempty"`

	// generated view current view

	InstallNumber            int                  `json:"install_number,omitzero" gorm:"->;-:migration" temporaljson:"install_number,omitzero,omitempty"`
	SandboxStatus            InstallSandboxStatus `json:"sandbox_status,omitzero" gorm:"->;-:migration" swaggertype:"string" temporaljson:"sandbox_status,omitzero,omitempty"`
	SandboxStatusDescription string               `json:"sandbox_status_description,omitzero" gorm:"-" swaggertype:"string" temporaljson:"sandbox_status_description,omitzero,omitempty"`
	ComponentStatuses        pgtype.Hstore        `json:"component_statuses,omitzero" gorm:"type:hstore;->;-:migration" swaggertype:"object,string" temporaljson:"component_statuses,omitzero,omitempty"`

	Workflows []Workflow `json:"workflows,omitzero" gorm:"polymorphic:Owner;constraint:OnDelete:CASCADE;" temporaljson:"workflows,omitzero,omitempty"`

	// after queries

	CurrentInstallInputs                *InstallInputs         `json:"-" gorm:"-" temporaljson:"current_install_inputs,omitzero,omitempty"`
	CompositeComponentStatus            InstallComponentStatus `json:"composite_component_status,omitzero" gorm:"-" swaggertype:"string" temporaljson:"composite_component_status,omitzero,omitempty"`
	CompositeComponentStatusDescription string                 `json:"composite_component_status_description,omitzero" gorm:"-" swaggertype:"string" temporaljson:"composite_component_status_description,omitzero,omitempty"`
	RunnerStatus                        RunnerStatus           `json:"runner_status,omitzero" gorm:"-" swaggertype:"string" temporaljson:"runner_status,omitzero,omitempty"`
	RunnerStatusDescription             string                 `json:"runner_status_description,omitzero" gorm:"-" swaggertype:"string" temporaljson:"runner_status_description,omitzero,omitempty"`
	RunnerID                            string                 `json:"runner_id,omitzero" gorm:"-" temporaljson:"runner_id,omitzero,omitempty"`
	CloudPlatform                       CloudPlatform          `json:"cloud_platform,omitzero" gorm:"-" swaggertype:"string" temporaljson:"cloud_platform,omitzero,omitempty"`
	RunnerType                          AppRunnerType          `json:"runner_type,omitzero" gorm:"-" swaggertype:"string" temporaljson:"runner_type,omitzero,omitempty"`

	Links map[string]any `json:"links,omitzero,omitempty" temporaljson:"-" gorm:"-"`

	// TODO(jm): deprecate these fields once the terraform provider has been updated
	Status            string `json:"status,omitzero" gorm:"-" temporaljson:"status,omitzero,omitempty"`
	StatusDescription string `json:"status_description,omitzero" gorm:"-" temporaljson:"status_description,omitzero,omitempty"`
}

func (i *Install) UseView() bool {
	return true
}

func (i *Install) ViewVersion() string {
	return "v6"
}

func (i *Install) Views(db *gorm.DB) []migrations.View {
	return []migrations.View{
		{
			Name:          views.DefaultViewName(db, &Install{}, 6),
			SQL:           viewsql.InstallsViewV6,
			AlwaysReapply: true,
		},
		{
			Name: views.CustomViewName(db, &Install{}, "migration_test"),
			SQL:  `SELECT * FROM installs ORDER BY created_at DESC`,
		},
	}
}

func (i *Install) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &Install{}, "migration_test"),
			Columns: []string{
				"org_id",
			},
		},
	}
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
	i.Links = links.InstallLinks(tx.Statement.Context, i.ID)

	// get the runner status
	i.RunnerStatus = RunnerStatusDeprovisioned
	if len(i.RunnerGroup.Runners) > 0 {
		i.RunnerStatus = i.RunnerGroup.Runners[0].Status
		i.RunnerStatusDescription = i.RunnerGroup.Runners[0].StatusDescription
		i.RunnerID = i.RunnerGroup.Runners[0].ID
	}

	if len(i.InstallInputs) > 0 {
		i.CurrentInstallInputs = &i.InstallInputs[0]
	}

	// get the composite status of all the components
	i.CompositeComponentStatus = compositeComponentStatus(i.ComponentStatuses)
	i.CompositeComponentStatusDescription = compositeComponentStatusDescription(i.ComponentStatuses)

	i.Status = "deprecated"
	i.StatusDescription = "deprecated, please use individual runner, sandbox and component statuses instead"

	if i.AppRunnerConfig.ID != "" {
		i.CloudPlatform = i.AppRunnerConfig.CloudPlatform
		i.RunnerType = i.AppRunnerConfig.Type

	} else {
		i.CloudPlatform = CloudPlatformUnknown
		i.RunnerType = AppRunnerTypeUnknown
	}

	return nil
}

// compositeComponentStatus coalesces a single status from the statuses of the app's components.
// This is based on the components defined in the app config, not the components present in the install.
// Components may be present in an install's history that have been removed from the app.
func compositeComponentStatus(componentStatuses pgtype.Hstore) InstallComponentStatus {
	// if there are no components, then there are no operations to wait for
	if len(componentStatuses) == 0 {
		return InstallComponentStatusPending
	}

	// check status of each component
	activecount := 0
	for _, status := range componentStatuses {
		switch InstallComponentStatus(*status) {
		case InstallComponentStatusActive:
			activecount++
		case InstallComponentStatusError:
			// if any components have failed, composite status should be "error"
			// we can stop immediately
			return InstallComponentStatusError
		}
	}

	// if all components are active, composite status should be "active"
	if activecount == len(componentStatuses) {
		return InstallComponentStatusActive
	}

	// if any components have not yet succeeded or failed, composite status should be "pending"
	return InstallComponentStatusPending
}

func compositeComponentStatusDescription(componentStatuses pgtype.Hstore) string {
	// if there are no components, then there are no operations to wait for
	if len(componentStatuses) == 0 {
		return "No active components"
	}

	// check status of each component
	activecount := 0
	for _, status := range componentStatuses {
		switch InstallComponentStatus(*status) {
		case InstallComponentStatusActive:
			activecount++
		case InstallComponentStatusError:
			// if any components have failed we can stop immediately
			return "A component is in an error state"
		}
	}

	// if all components are active
	if activecount == len(componentStatuses) {
		return "All components have been deployed"
	}

	// if any components have not yet succeeded or failed
	return "Waiting on components"
}

func (i *Install) EventLoops() []bulk.EventLoop {
	evs := make([]bulk.EventLoop, 0)
	evs = append(evs, bulk.EventLoop{
		Namespace: "installs",
		ID:        i.ID,
	})

	for _, runner := range i.RunnerGroup.Runners {
		evs = append(evs, bulk.EventLoop{
			Namespace: "runners",
			ID:        runner.ID,
		})
	}

	return evs
}
