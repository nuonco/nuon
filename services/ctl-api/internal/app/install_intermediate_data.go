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
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"not null;default:null"`
	CreatedBy   Account               `json:"-"`
	CreatedAt   time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time             `json:"updated_at" gorm:"notnull"`
	DeletedAt   soft_delete.DeletedAt `json:"-"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`
	Org   Org    `json:"-" faker:"-"`

	InstallID string  `json:"install_id" gorm:"not null;default null"`
	Install   Install `swaggerignore:"true" json:"-"`

	RunnerJob   RunnerJob `json:"-"`
	RunnerJobID string    `json:"runner_job_id"`

	IntermediateDataJSON string `json:"-" temporaljson:"intermediate_data_json" gorm:"default null;not null"`

	// loaded via after query

	IntermediateData map[string]interface{} `json:"intermediate_data" gorm:"-"`
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
