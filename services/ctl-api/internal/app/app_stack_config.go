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
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	AppID       string `json:"app_id" temporaljson:"app_id,omitzero,omitempty"`
	AppConfigID string `json:"app_config_id" temporaljson:"app_config_id,omitzero,omitempty"`

	Type                    StackType `json:"type" temporaljson:"type,omitzero,omitempty"`
	Name                    string    `json:"name" features:"template" temporaljson:"name,omitzero,omitempty"`
	Description             string    `json:"description" features:"template" temporaljson:"description,omitzero,omitempty"`
	RunnerNestedTemplateURL string    `json:"runner_nested_template_url" temporaljson:"runner_nested_template_url,omitzero,omitempty"`
	VPCNestedTemplateURL    string    `json:"vpc_nested_template_url" temporaljson:"vpc_nested_template_url,omitzero,omitempty"`
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
