package app

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

// TestGroup represents a group of canary tests that are run together.
// It is associated with a single Canary and would typically represent a single nuon application.
type TestGroup struct {
	ID        string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedAt time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"index:idx_app_name,unique" temporaljson:"deleted_at,omitzero,omitempty"`

	CanaryID string  `json:"canary_id,omitzero" gorm:"not null;index" temporaljson:"canary_id,omitzero,omitempty"`
	Canary   TestRun `faker:"-" json:"-" temporaljson:"canary,omitzero,omitempty"`

	AppConfigPath    string `json:"app_config_path,omitzero" gorm:"not null" temporaljson:"app_config_path,omitzero,omitempty"`
	AppConfigGithash string `json:"app_config_githash,omitzero" gorm:"not null" temporaljson:"app_config_githash,omitzero,omitempty"`

	StartTime time.Time `gorm:"not null;index:idx_test_name_time,priority:2" json:"start_time"`
	EndTime   time.Time `json:"end_time"`

	TestExecution []TestExecution `faker:"-" json:"test_executions,omitempty" temporaljson:"test_executions,omitempty"`
}
