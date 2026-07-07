package app

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type JSONMap map[string]string

type HelmRelease struct {
	HelmChartID string    `gorm:"primary_key:true"`
	HelmChart   HelmChart `gorm:"-" temporaljson:"helm_chart,omitzero,omitempty"`

	CreatedByID string  `json:"created_by_id" gorm:"not null;default:null" `
	CreatedBy   Account `json:"-"`

	CreatedAt time.Time             `json:"created_at" gorm:"notnull"`
	UpdatedAt time.Time             `json:"updated_at" gorm:"notnull" `
	DeletedAt soft_delete.DeletedAt `json:"-" `

	OrgID string `json:"org_id" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" temporaljson:"org,omitzero,omitempty"`

	Key string `gorm:"primaryKey:true"`

	// See https://github.com/helm/helm/blob/c9fe3d118caec699eb2565df9838673af379ce12/pkg/storage/driver/secrets.go#L231
	Type string `gorm:"not null"`

	// The rspb.Release body (base64-encoded), stored in S3 via blob storage.
	Body *blobstore.Blob `gorm:"column:body_blob" json:"-" temporaljson:"-"`

	// Release "labels" that can be used as filters in the storage.Query(labels map[string]string)
	// we implemented. Note that allowing Helm users to filter against new dimensions will require a
	// new migration to be added, and the Create and/or update functions to be updated accordingly.
	Name      string `gorm:"not null"`
	Namespace string `gorm:"not null"`
	Version   int    `gorm:"not null"`
	Status    string `gorm:"not null"`
	Owner     string `gorm:"not null"`

	Labels JSONMap `json:"labels,omitempty"`
}

func (t *HelmRelease) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &HelmRelease{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (t *HelmRelease) BeforeCreate(tx *gorm.DB) (err error) {
	if t.CreatedByID == "" {
		t.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	}

	if t.OrgID == "" {
		t.OrgID = orgIDFromContext(tx.Statement.Context)
	}
	return nil
}

// BeforeSave (not BeforeCreate) so the body blob uploads on helm upgrades too, which use Updates.
func (t *HelmRelease) BeforeSave(tx *gorm.DB) error {
	return t.Body.BeforeCreate(tx)
}

func (j JSONMap) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *JSONMap) Scan(value any) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("invalid type for JSONMap")
	}
	return json.Unmarshal(bytes, &j)
}
