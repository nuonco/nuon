package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop/bulk"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/links"
)

type AppStatus string

const (
	AppStatusProvisioning   AppStatus = "provisioning"
	AppStatusDeprovisioning AppStatus = "deprovisioning"
	AppStatusActive         AppStatus = "active"
	AppStatusError          AppStatus = "error"
)

type App struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id" gorm:"notnull" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index:idx_app_name,unique" temporaljson:"deleted_at,omitzero,omitempty"`

	Name        string              `json:"name" gorm:"index:idx_app_name,unique" temporaljson:"name,omitzero,omitempty"`
	Description generics.NullString `json:"description" swaggertype:"string" temporaljson:"description,omitzero,omitempty"`
	DisplayName generics.NullString `json:"display_name" swaggertype:"string" temporaljson:"display_name,omitzero,omitempty"`

	OrgID string `json:"org_id" gorm:"index:idx_app_name,unique" temporaljson:"org_id,omitzero,omitempty"`
	Org   *Org   `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	NotificationsConfig NotificationsConfig `gorm:"polymorphic:Owner;constraint:OnDelete:CASCADE;" json:"notifications_config,omitempty" temporaljson:"notifications_config,omitzero,omitempty"`

	Components                 []Component        `faker:"components" json:"-" swaggerignore:"true" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"components,omitzero,omitempty"`
	Installs                   []Install          `faker:"-" json:"-" swaggerignore:"true" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"installs,omitzero,omitempty"`
	ActionWorkflows            []ActionWorkflow   `json:"-" swaggerignore:"true" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"action_workflows,omitzero,omitempty"`
	AppInputConfigs            []AppInputConfig   `json:"-" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"app_input_configs,omitzero,omitempty"`
	AppSandboxConfigs          []AppSandboxConfig `json:"-" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"app_sandbox_configs,omitzero,omitempty"`
	AppRunnerConfigs           []AppRunnerConfig  `json:"-" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"app_runner_configs,omitzero,omitempty"`
	CloudFormationStackConfigs []AppStackConfig   `json:"-" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"cloud_formation_stack_configs,omitzero,omitempty"`
	AppConfigs                 []AppConfig        `json:"-" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"app_configs,omitzero,omitempty"`
	AppSecrets                 []AppSecret        `json:"-" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"app_secrets,omitzero,omitempty"`
	InstallerApps              []InstallerApp     `json:"-" gorm:"constraint:OnDelete:CASCADE;" temporaljson:"installer_apps,omitzero,omitempty"`

	Status            AppStatus `json:"status" swaggertype:"string" temporaljson:"status,omitzero,omitempty"`
	StatusDescription string    `json:"status_description" temporaljson:"status_description,omitzero,omitempty"`

	// fields set via after query
	AppInputConfig   AppInputConfig   `json:"input_config" gorm:"-" temporaljson:"app_input_config,omitzero,omitempty"`
	AppSandboxConfig AppSandboxConfig `json:"sandbox_config" gorm:"-" temporaljson:"app_sandbox_config,omitzero,omitempty"`
	AppRunnerConfig  AppRunnerConfig  `json:"runner_config" gorm:"-" temporaljson:"app_runner_config,omitzero,omitempty"`

	Links map[string]any `json:"links,omitempty" temporaljson:"-" gorm:"-"`

	CloudPlatform CloudPlatform `json:"cloud_platform" gorm:"-" swaggertype:"string" temporaljson:"cloud_platform,omitzero,omitempty"`
}

func (a *App) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAppID()
	}

	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	return nil
}

func (a *App) AfterQuery(tx *gorm.DB) error {
	cfg := configFromContext(tx.Statement.Context)
	if cfg != nil {
		a.Links = links.AppLinks(cfg, a.ID)
	}

	a.CloudPlatform = CloudPlatformUnknown
	if len(a.AppRunnerConfigs) > 0 {
		a.AppRunnerConfig = a.AppRunnerConfigs[0]
		a.CloudPlatform = a.AppRunnerConfigs[0].CloudPlatform
	}
	if len(a.AppInputConfigs) > 0 {
		a.AppInputConfig = a.AppInputConfigs[0]
	}
	if len(a.AppSandboxConfigs) > 0 {
		a.AppSandboxConfig = a.AppSandboxConfigs[0]
	}

	return nil
}

func (a *App) EventLoops() []bulk.EventLoop {
	evs := make([]bulk.EventLoop, 0)
	evs = append(evs, bulk.EventLoop{
		Namespace: "apps",
		ID:        a.ID,
	})

	for _, cmp := range a.Components {
		evs = append(evs, bulk.EventLoop{
			Namespace: "components",
			ID:        cmp.ID,
		})
	}

	for _, acw := range a.ActionWorkflows {
		evs = append(evs, bulk.EventLoop{
			Namespace: "actions",
			ID:        acw.ID,
		})
	}

	for _, inst := range a.Installs {
		evs = append(evs, inst.EventLoops()...)
	}

	return evs
}
