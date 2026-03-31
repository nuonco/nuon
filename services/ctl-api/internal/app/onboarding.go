package app

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/lib/pq"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/shortid/domains"
)

// IntermediateAppConfig is a custom JSONB type that stores a config.AppConfig in PostgreSQL.
type IntermediateAppConfig struct {
	config.AppConfig
}

func (c *IntermediateAppConfig) Scan(v interface{}) (err error) {
	switch v := v.(type) {
	case nil:
		return nil
	case []byte:
		if err := json.Unmarshal(v, &c.AppConfig); err != nil {
			return errors.Wrap(err, "unable to scan intermediate app config")
		}
	}
	return
}

func (c *IntermediateAppConfig) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c.AppConfig)
}

func (IntermediateAppConfig) GormDataType() string {
	return "jsonb"
}

type OnboardingStatus string

const (
	OnboardingStatusActive    OnboardingStatus = "active"
	OnboardingStatusCompleted OnboardingStatus = "completed"
	OnboardingStatusAbandoned OnboardingStatus = "abandoned"
)

type OnboardingStep string

const (
	OnboardingStepOrganization OnboardingStep = "organization"
	OnboardingStepYourStack    OnboardingStep = "your_stack"
	OnboardingStepInstall      OnboardingStep = "install"
	OnboardingStepDeploy       OnboardingStep = "deploy"
	OnboardingStepGetStarted   OnboardingStep = "get_started"
)

type OnboardingStepStatus string

const (
	OnboardingStepStatusActive     OnboardingStepStatus = "active"
	OnboardingStepStatusProcessing OnboardingStepStatus = "processing"
	OnboardingStepStatusError      OnboardingStepStatus = "error"
)

type OnboardingAppType string

const (
	OnboardingAppTypeCustom  OnboardingAppType = "custom"
	OnboardingAppTypeExample OnboardingAppType = "example"
)

type OnboardingInstallMode string

const (
	OnboardingInstallModeCloud   OnboardingInstallMode = "cloud"
	OnboardingInstallModeSandbox OnboardingInstallMode = "sandbox"
)

type Onboarding struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"-"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `json:"-"`

	AccountID string  `json:"account_id,omitzero" gorm:"not null;index"`
	Account   Account `json:"-"`

	Status      OnboardingStatus `json:"status,omitzero" gorm:"notnull" swaggertype:"string"`
	StatusV2    CompositeStatus  `json:"status_v2,omitzero" gorm:"type:jsonb" temporaljson:"status_v2,omitzero,omitempty"`
	CurrentStep OnboardingStep   `json:"current_step,omitzero" gorm:"notnull" swaggertype:"string"`

	// Step 1: Organization
	OrgID *string `json:"org_id,omitempty"`
	Org   *Org    `json:"-"`

	// Step 2: Your Stack
	AppType        OnboardingAppType      `json:"app_type,omitempty" swaggertype:"string"`
	ExampleAppSlug *string                `json:"example_app_slug,omitempty"`
	CloudProvider  *string                `json:"cloud_provider,omitempty"`
	AppAttributes  pq.StringArray         `json:"app_attributes,omitempty" gorm:"type:text[];default '{}'" swaggertype:"array,string"`
	AppConfig      *IntermediateAppConfig `json:"app_config,omitempty" swaggertype:"object"`
	AppID          *string                `json:"app_id,omitempty"`
	App            *App                   `json:"-"`
	AppBranchID    *string                `json:"app_branch_id,omitempty"`
	AppBranch      *AppBranch             `json:"-"`

	// Step 3: Install
	InstallMode OnboardingInstallMode `json:"install_mode,omitempty" swaggertype:"string"`
	InstallID   *string               `json:"install_id,omitempty"`
	Install     *Install              `json:"-"`
	WorkflowID  *string               `json:"workflow_id,omitempty"`

	// Async step status (for queue-based signal processing)
	StepStatus OnboardingStepStatus `json:"step_status,omitempty" gorm:"default:''" swaggertype:"string"`
	StepError  *string              `json:"step_error,omitempty"`
}

func (o *Onboarding) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = domains.NewOnboardingID()
	}
	o.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	return nil
}

func (o *Onboarding) SetCompositeStatus(ctx context.Context, status Status) {
	existing := o.StatusV2
	existing.History = nil
	newStatus := NewCompositeStatus(ctx, status)
	newStatus.History = append(o.StatusV2.History, existing)
	o.StatusV2 = newStatus
}
