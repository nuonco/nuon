package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type AppPolicyType string

const (
	AppPolicyTypeKubernetesClusterKyverno string = "kubernetes_cluster"

	AppPolicyTypeTerraformDeployRunnerJobKyverno string = "runner_job_terraform_deploy"
	AppPolicyTypeHelmDeployRunnerJobKyverno      string = "runner_job_helm_deploy"
	AppPolicyTypeActionWorkflowRunnerJobKyverno  string = "runner_job_action_workflow"
)

type AppPolicyConfig struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"-"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `json:"-"`

	OrgID       string `json:"org_id" gorm:"notnull;default null"`
	AppID       string `json:"app_id" gorm:"notnull;default null"`
	AppConfigID string `json:"app_config_id" gorm:"notnull;default null"`

	AppPoliciesConfigID string            `json:"app_policies_config" gorm:"notnull;default null"`
	AppPoliciesConfig   AppPoliciesConfig `json:"-"`

	Type     AppPolicyType `json:"type"`
	Contents string        `json:"contents"`
}

func (a *AppPolicyConfig) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAppCfgID()
	}
	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
	}

	return nil
}
