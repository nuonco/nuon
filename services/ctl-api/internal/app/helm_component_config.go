package app

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type HelmComponentConfig struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	// parent reference
	ComponentConfigConnectionID string                    `json:"component_config_connection_id,omitzero" gorm:"notnull" temporaljson:"component_config_connection_id,omitzero,omitempty"`
	ComponentConfigConnection   ComponentConfigConnection `json:"-" temporaljson:"component_config_connection,omitzero,omitempty"`

	HelmConfig *HelmConfig `json:"helm_config_json,omitzero" gorm:"type:jsonb" temporaljson:"helm_config_json,omitzero,omitempty"`

	// Helm specific configurations
	ChartName     string              `json:"chart_name,omitzero" gorm:"notnull" features:"template" temporaljson:"chart_name,omitzero,omitempty"`
	Values        pgtype.Hstore       `json:"values,omitzero" gorm:"type:hstore" swaggertype:"object,string" features:"template" temporaljson:"values,omitzero,omitempty"`
	ValuesFiles   pq.StringArray      `gorm:"type:text[]" json:"values_files,omitzero" swaggertype:"array,string" features:"template" temporaljson:"values_files,omitzero,omitempty"`
	Namespace     generics.NullString `json:"namespace,omitzero" swaggertype:"string" features:"template" temporaljson:"namespace,omitzero,omitempty"`
	StorageDriver generics.NullString `json:"storage_driver,omitzero" swaggertype:"string" features:"template" temporaljson:"storage_driver,omitzero,omitempty"`
	// Newer config fields that we don't need a column for
	TakeOwnership bool `json:"take_ownership,omitzero" gorm:"-" temporaljson:"take_ownership,omitzero,omitempty"`

	PublicGitVCSConfig       *PublicGitVCSConfig       `gorm:"polymorphic:ComponentConfig;constraint:OnDelete:CASCADE;" json:"public_git_vcs_config,omitzero,omitempty" temporaljson:"public_git_vcs_config,omitzero,omitempty"`
	ConnectedGithubVCSConfig *ConnectedGithubVCSConfig `gorm:"polymorphic:ComponentConfig;constraint:OnDelete:CASCADE;" json:"connected_github_vcs_config,omitzero,omitempty" temporaljson:"connected_github_vcs_config,omitzero,omitempty"`
}

func (c *HelmComponentConfig) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &HelmComponentConfig{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (c *HelmComponentConfig) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewConfigID()
	c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	c.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}

func (c *HelmComponentConfig) AfterQuery(tx *gorm.DB) error {
	if c.HelmConfig != nil {
		c.ChartName = c.HelmConfig.ChartName
		c.Values = c.HelmConfig.Values
		c.ValuesFiles = c.HelmConfig.ValuesFiles
		c.Namespace = generics.NewNullString(c.HelmConfig.Namespace)
		c.StorageDriver = generics.NewNullString(c.HelmConfig.StorageDriver)
		c.TakeOwnership = c.HelmConfig.TakeOwnership
	}
	return nil
}

type HelmConfig struct {
	ChartName      string             `json:"chart_name"`
	Values         map[string]*string `json:"values"`
	ValuesFiles    []string           `json:"values_files"`
	Namespace      string             `json:"namespace"`
	StorageDriver  string             `json:"storage_driver"`
	HelmRepoConfig *HelmRepoConfig    `json:"helm_repo_config,omitempty"`
	// Newer fields that we don't need to store as columns in the database
	TakeOwnership bool `json:"take_ownership,omitempty"`
}

type HelmRepoConfig struct {
	RepoURL string `json:"repo_url"`
	Chart   string `json:"chart"`
	Version string `json:"version,omitempty"`
}

// Scan implements the database/sql.Scanner interface.
func (c *HelmConfig) Scan(v interface{}) (err error) {
	switch v := v.(type) {
	case nil:
		return nil
	case []byte:
		if err := json.Unmarshal(v, c); err != nil {
			return errors.Wrap(err, "unable to scan helm config")
		}
	}
	return
}

// Value implements the driver.Valuer interface.
func (c *HelmConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

func (HelmConfig) GormDataType() string {
	return "jsonb"
}
