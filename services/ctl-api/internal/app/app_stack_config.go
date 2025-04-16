package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type StackType string

const (
	StackTypeAWS StackType = "aws-cloudformation"
)

type AppStackConfig struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"-"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `json:"-"`

	OrgID string `json:"org_id" gorm:"notnull;default null"`
	Org   Org    `faker:"-" json:"-"`

	AppID       string `json:"app_id"`
	AppConfigID string `json:"app_config_id"`

	Type                    StackType `json:"type"`
	Name                    string    `json:"name" features:"template"`
	Description             string    `json:"description" features:"template"`
	RunnerNestedTemplateURL string    `json:"runner_nested_template_url"`
	VPCNestedTemplateURL    string    `json:"vpc_nested_template_url"`
}

func (a *AppStackConfig) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewAppID()
	}
	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	if a.OrgID == "" {
		a.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}
