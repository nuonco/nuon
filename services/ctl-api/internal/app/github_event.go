package app

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type GithubEvent struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	GithubInstallID string           `json:"github_install_id,omitempty" gorm:"default:null" temporaljson:"github_install_id,omitzero,omitempty"`
	EventType       string           `json:"event_type" gorm:"not null;default:''" temporaljson:"event_type,omitzero,omitempty"`
	Payload         *blobstore.Blob  `json:"payload" gorm:"type:jsonb;default:null" temporaljson:"payload,omitzero,omitempty"`
	Status          *CompositeStatus `json:"status,omitempty" gorm:"column:status;type:jsonb;default:null" temporaljson:"status,omitzero,omitempty"`
}

func (g *GithubEvent) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name:    indexes.Name(db, &GithubEvent{}, "github_install_id"),
			Columns: []string{"github_install_id"},
		},
	}
}

func (g *GithubEvent) BeforeCreate(tx *gorm.DB) error {
	g.ID = domains.NewGithubEventID()
	if g.CreatedByID == "" {
		g.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	if g.Payload != nil {
		if err := g.Payload.BeforeCreate(tx); err != nil {
			return err
		}
	}

	return nil
}
