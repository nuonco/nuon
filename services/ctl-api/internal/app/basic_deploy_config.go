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
	CreatedByID string         `json:"created_by_id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// connection to parent model
	ComponentConfigID   string `json:"component_config_id"`
	ComponentConfigType string `json:"component_config_type"`

	// actual configuration
	InstanceCount   int            `json:"instance_count"`
	ListenPort      int            `json:"listen_port"`
	HealthCheckPath string         `json:"health_check_path"`
	CPURequest      string         `json:"cpu_request"`
	CPULimit        string         `json:"cpu_limit"`
	MemRequest      string         `json:"mem_request"`
	MemLimit        string         `json:"mem_limit"`
	EnvVars         pgtype.Hstore  `json:"env_vars" gorm:"type:hstore" swaggertype:"object,string"`
	Args            pq.StringArray `gorm:"type:text[]" json:"args" swaggertype:"array,string"`
}

func (c *BasicDeployConfig) BeforeCreate(tx *gorm.DB) error {
	c.ID = domains.NewComponentID()
	return nil
}
