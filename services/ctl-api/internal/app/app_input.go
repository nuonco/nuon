package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/iancoleman/strcase"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type AppInputType string

const (
	AppInputTypeString AppInputType = "string"
	AppInputTypeNumber AppInputType = "number"
	AppInputTypeBool   AppInputType = "bool"
	AppInputTypeList   AppInputType = "list"
	AppInputTypeJSON   AppInputType = "json"
)

type AppInputSource string

const (
	AppInputSourceVendor   AppInputSource = "vendor"
	AppInputSourceCustomer AppInputSource = "customer"
)

type AppInput struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" gorm:"index:idx_app_input_unique_name,unique" temporaljson:"deleted_at,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" gorm:"notnull;default null" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" temporaljson:"org,omitzero,omitempty"`

	AppInputConfigID string         `json:"app_input_id,omitzero" gorm:"notnull; default null;index:idx_app_input_unique_name,unique" temporaljson:"app_input_config_id,omitzero,omitempty"`
	AppInputConfig   AppInputConfig `json:"-" temporaljson:"app_input_config,omitzero,omitempty"`

	AppInputGroup   AppInputGroup `json:"group,omitzero" temporaljson:"app_input_group,omitzero,omitempty"`
	AppInputGroupID string        `json:"group_id,omitzero" temporaljson:"app_input_group_id,omitzero,omitempty"`

	Name        string `json:"name,omitzero" gorm:"not null;default null;index:idx_app_input_unique_name,unique" temporaljson:"name,omitzero,omitempty"`
	DisplayName string `json:"display_name,omitzero" temporaljson:"display_name,omitzero,omitempty"`
	Description string `json:"description,omitzero" gorm:"not null; default null" temporaljson:"description,omitzero,omitempty"`
	Default     string `json:"default,omitzero" temporaljson:"default,omitzero,omitempty"`
	Required    bool   `json:"required,omitzero" temporaljson:"required,omitzero,omitempty"`
	Sensitive   bool   `json:"sensitive,omitzero" temporaljson:"sensitive,omitzero,omitempty"`

	Index    int            `json:"index,omitzero"`
	Internal bool           `json:"internal,omitzero"`
	Type     AppInputType   `json:"type,omitzero" swaggertype:"string"`
	Source   AppInputSource `json:"source,omitzero" gorm:"not null;default:'vendor'" swaggertype:"string" temporaljson:"source"`

	// CloudFormation configuration (computed fields, not stored in DB)
	CloudFormationStackName      string `json:"cloudformation_stack_name,omitzero" gorm:"-" temporaljson:"cloudformation_stack_name,omitzero,omitempty"`
	CloudFormationStackParamName string `json:"cloudformation_stack_parameter_name,omitzero" gorm:"-" temporaljson:"cloudformation_stack_parameter_name,omitzero,omitempty"`
}

func (a *AppInput) BeforeCreate(tx *gorm.DB) error {
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

func (a *AppInput) AfterQuery(tx *gorm.DB) error {
	// Compute CloudFormation configuration fields for install_stack sourced inputs
	if a.Source == AppInputSourceCustomer {
		a.CloudFormationStackName = computeCloudFormationStackName(a.Name)
		a.CloudFormationStackParamName = computeCloudFormationStackParameterName(a.Name)
	}
	return nil
}

// computeCloudFormationStackName generates the CloudFormation stack resource name
func computeCloudFormationStackName(inputName string) string {
	return "Install" + strcase.ToCamel(inputName)
}

// computeCloudFormationStackParameterName generates the CloudFormation parameter name
func computeCloudFormationStackParameterName(inputName string) string {
	return "Install" + strcase.ToCamel(inputName)
}
