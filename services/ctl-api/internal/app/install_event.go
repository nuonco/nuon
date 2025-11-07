package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type OperationStatus string

const (
	OperationStatusStarted  OperationStatus = "started"
	OperationStatusFinished OperationStatus = "finished"
	OperationStatusNoop     OperationStatus = "noop"
	OperationStatusFailed   OperationStatus = "failed"
)

type InstallEvent struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"notnull" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	InstallID string  `json:"install_id,omitzero" temporaljson:"install_id,omitzero,omitempty"`
	Install   Install `swaggerignore:"-" json:"-" temporaljson:"install,omitzero,omitempty"`

	OrgID string `json:"org_id,omitzero" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `faker:"-" json:"-" swaggerignore:"-" temporaljson:"org,omitzero,omitempty"`

	Operation       string          `json:"operation,omitzero" temporaljson:"operation,omitzero,omitempty"`
	OperationStatus OperationStatus `json:"operation_status,omitzero" temporaljson:"operation_status,omitzero,omitempty"`

	Payload []byte `json:"payload,omitzero" gorm:"type:jsonb" swaggertype:"object,string" temporaljson:"payload,omitzero,omitempty"`

	OperationName string `gorm:"-" json:"operation_name,omitzero" temporaljson:"operation_name,omitzero,omitempty"`
}

func (i *InstallEvent) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &InstallEvent{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (a *InstallEvent) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = domains.NewEventID()
	}

	if a.CreatedByID == "" {
		a.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}
	return nil
}

func (i *InstallEvent) AfterQuery(tx *gorm.DB) error {
	i.OperationName = generics.DisplayName(i.Operation)
	return nil
}
