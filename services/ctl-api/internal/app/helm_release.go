package app

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
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

	// The rspb.Release body, as a base64-encoded string
	Body string `gorm:"not null"`

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
