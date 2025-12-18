package app

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/indexes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
)

type KubernetesManifestComponentConfig struct {
	ID          string                `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedByID string                `json:"created_by_id,omitzero" gorm:"not null;default:null" temporaljson:"created_by_id,omitzero,omitempty"`
	CreatedBy   Account               `json:"-" temporaljson:"created_by,omitzero,omitempty"`
	CreatedAt   time.Time             `json:"created_at,omitzero" gorm:"notnull" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt   time.Time             `json:"updated_at,omitzero" gorm:"notnull" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt   soft_delete.DeletedAt `gorm:"index" json:"-" temporaljson:"deleted_at,omitzero,omitempty"`

	// used for RLS
	OrgID string `json:"org_id,omitzero" gorm:"notnull" swaggerignore:"true" temporaljson:"org_id,omitzero,omitempty"`
	Org   Org    `json:"-" faker:"-" temporaljson:"org,omitzero,omitempty"`

	// value
	ComponentConfigConnectionID string                    `json:"component_config_connection_id,omitzero" gorm:"notnull" temporaljson:"component_config_connection_id,omitzero,omitempty"`
	ComponentConfigConnection   ComponentConfigConnection `json:"-" temporaljson:"component_config_connection,omitzero,omitempty"`

	// Primary fields - used for inline manifests (fully supported)
	Manifest  string `json:"manifest,omitzero" gorm:"not null;default:''" temporaljson:"manifest,omitzero,omitempty"`
	Namespace string `json:"namespace,omitzero" gorm:"not null;default:default" temporaljson:"namespace,omitzero,omitempty"`

	// Kustomize configuration (mutually exclusive with Manifest)
	Kustomize *KustomizeConfig `json:"kustomize,omitzero" gorm:"type:jsonb" temporaljson:"kustomize,omitzero,omitempty"`

	// VCS configuration for kustomize sources (similar to HelmComponentConfig)
	PublicGitVCSConfig       *PublicGitVCSConfig       `gorm:"polymorphic:ComponentConfig;constraint:OnDelete:CASCADE;" json:"public_git_vcs_config,omitzero,omitempty" temporaljson:"public_git_vcs_config,omitzero,omitempty"`
	ConnectedGithubVCSConfig *ConnectedGithubVCSConfig `gorm:"polymorphic:ComponentConfig;constraint:OnDelete:CASCADE;" json:"connected_github_vcs_config,omitzero,omitempty" temporaljson:"connected_github_vcs_config,omitzero,omitempty"`
}

// KustomizeConfig defines kustomize build options
type KustomizeConfig struct {
	// Path to kustomization directory (relative to source root)
	Path string `json:"path"`

	// Additional patch files to apply after kustomize build
	Patches []string `json:"patches,omitempty"`

	// Enable Helm chart inflation during kustomize build
	EnableHelm bool `json:"enable_helm,omitempty"`

	// Load restrictor: "none" or "rootOnly" (default: "rootOnly")
	LoadRestrictor string `json:"load_restrictor,omitempty"`
}

// Scan implements the database/sql.Scanner interface
func (c *KustomizeConfig) Scan(v interface{}) (err error) {
	switch v := v.(type) {
	case nil:
		return nil
	case []byte:
		if err := json.Unmarshal(v, c); err != nil {
			return errors.Wrap(err, "unable to scan kustomize config")
		}
	}
	return
}

// Value implements the driver.Valuer interface
func (c *KustomizeConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// GormDataType returns the GORM data type for this field
func (KustomizeConfig) GormDataType() string {
	return "jsonb"
}

// SourceType returns the source type based on which fields are populated
func (k *KubernetesManifestComponentConfig) SourceType() string {
	if k.Kustomize != nil && k.Kustomize.Path != "" {
		return "kustomize"
	}
	return "inline"
}

func (k *KubernetesManifestComponentConfig) Indexes(db *gorm.DB) []migrations.Index {
	return []migrations.Index{
		{
			Name: indexes.Name(db, &KubernetesManifestComponentConfig{}, "org_id"),
			Columns: []string{
				"org_id",
			},
		},
	}
}

func (e *KubernetesManifestComponentConfig) BeforeCreate(tx *gorm.DB) error {
	e.ID = domains.NewComponentID()
	e.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	e.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
