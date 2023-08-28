package app

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/lib/pq"
	"github.com/powertoolsdev/mono/pkg/shortid/domains"
	"gorm.io/gorm"
)

type BasicDeployConfig struct {
	ID          string         `gorm:"primary_key;check:id_checker,char_length(id)=26" json:"id"`
	CreatedByID string         `json:"created_by_id" gorm:"notnull"`
	CreatedAt   time.Time      `json:"created_at" gorm:"notnull"`
	UpdatedAt   time.Time      `json:"updated_at" gorm:"notnull"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// used for RLS
	OrgID string `json:"org_id" gorm:"notnull" swaggerignore:"true"`

	// connection to parent model
	ComponentConfigID   string `json:"component_config_id" gorm:"notnull"`
	ComponentConfigType string `json:"component_config_type" gorm:"notnull"`

	// actual configuration
	InstanceCount   int            `json:"instance_count" gorm:"notnull"`
	ListenPort      int            `json:"listen_port" gorm:"notnull"`
	HealthCheckPath string         `json:"health_check_path" gorm:"notnull"`
	CPURequest      string         `json:"cpu_request" gorm:"notnull"`
	CPULimit        string         `json:"cpu_limit" gorm:"notnull"`
	MemRequest      string         `json:"mem_request" gorm:"notnull"`
	MemLimit        string         `json:"mem_limit" gorm:"notnull"`
	EnvVars         pgtype.Hstore  `json:"env_vars" gorm:"type:hstore" swaggertype:"object,string"`
	Args            pq.StringArray `gorm:"type:text[]" json:"args" swaggertype:"array,string"`
}

func (c *BasicDeployConfig) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewComponentID()
	c.CreatedByID = createdByIDFromContext(tx.Statement.Context)
	c.OrgID = orgIDFromContext(tx.Statement.Context)
	return nil
}
