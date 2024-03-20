package app

import (
	"time"

	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/signals"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type OperationStatus string

const (
	OperationStatusStarted  OperationStatus = "started"
	OperationStatusFinished OperationStatus = "finished"
	OperationStatusNoop     OperationStatus = "noop"
	OperationStatusFailed   OperationStatus = "failed"
)

type InstallEvent struct {
	ID          string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string                `json:"created_by_id" gorm:"notnull"`
	CreatedBy   UserToken             `json:"created_by" gorm:"references:Subject"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	DeletedAt   soft_delete.DeletedAt `json:"-"`

	InstallID string  `json:"install_id"`
	Install   Install `swaggerignore:"-" json:"-"`

	OrgID string `json:"org_id"`
	Org   Org    `faker:"-" json:"-" swaggerignore:"-"`

	Operation       signals.Operation `json:"operation"`
	OperationStatus OperationStatus   `json:"operation_status"`

	Payload []byte `json:"payload" gorm:"type:jsonb" swaggertype:"object,string"`
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
