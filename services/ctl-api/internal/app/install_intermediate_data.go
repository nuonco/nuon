package app

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type InstallIntermediateData struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	InstallID string  `json:"install_id,omitzero" gorm:"not null;default null" temporaljson:"install_id,omitzero,omitempty"`
	Install   Install `swaggerignore:"true" json:"-" temporaljson:"install,omitzero,omitempty"`

	RunnerJob   RunnerJob `json:"-" temporaljson:"runner_job,omitzero,omitempty"`
	RunnerJobID string    `json:"runner_job_id,omitzero" temporaljson:"runner_job_id,omitzero,omitempty"`

	IntermediateDataJSON string `json:"-" gorm:"default null;not null" temporaljson:"intermediate_data_json,omitzero,omitempty"`

	// loaded via after query

	IntermediateData map[string]interface{} `json:"intermediate_data,omitzero" gorm:"-" temporaljson:"intermediate_data,omitzero,omitempty"`
}

func (i *InstallIntermediateData) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &InstallIntermediateData{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
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
