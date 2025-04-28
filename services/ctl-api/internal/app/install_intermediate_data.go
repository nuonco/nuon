package app

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
)

type InstallIntermediateData struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	InstallID string  `json:"install_id" gorm:"not null;default null" temporaljson:"install_id,omitzero,omitempty"`
	Install   Install `swaggerignore:"true" json:"-" temporaljson:"install,omitzero,omitempty"`

	RunnerJob   RunnerJob `json:"-" temporaljson:"runner_job,omitzero,omitempty"`
	RunnerJobID string    `json:"runner_job_id" temporaljson:"runner_job_id,omitzero,omitempty"`

	IntermediateDataJSON string `json:"-" gorm:"default null;not null" temporaljson:"intermediate_data_json,omitzero,omitempty"`

	// loaded via after query

	IntermediateData map[string]interface{} `json:"intermediate_data" gorm:"-" temporaljson:"intermediate_data,omitzero,omitempty"`
}

func (i *InstallIntermediateData) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = domains.NewInstallID()
	}

	i.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	i.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}

func (i *InstallIntermediateData) AfterQuery(tx *gorm.DB) error {
	var intermediateData map[string]interface{}

	if err := json.Unmarshal([]byte(i.IntermediateDataJSON), &intermediateData); err != nil {
		return errors.Wrap(err, "unable to parse intermediate data")
	}

	i.IntermediateData = intermediateData

	return nil
}
