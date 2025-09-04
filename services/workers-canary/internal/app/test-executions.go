package app

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

type TestExecutionStatus string

const (
	TestExecutionStatusPassed  TestExecutionStatus = "passed"
	TestExecutionStatusFailed  TestExecutionStatus = "failed"
	TestExecutionStatusSkipped TestExecutionStatus = "skipped"
)

// TestExecution represents the execution result of a single canary test within a CanaryGroup.
type TestExecution struct {
	ID        string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedAt time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"index:idx_app_name,unique" temporaljson:"deleted_at,omitzero,omitempty"`

	CanaryGroupID string    `json:"canary_group_id,omitzero" gorm:"not null;index" temporaljson:"canary_run_id,omitzero,omitempty"`
	CanaryGroup   TestGroup `faker:"-" json:"-" temporaljson:"canary_run,omitzero,omitempty"`

	TestName        string              `gorm:"not null;index:idx_test_name_time,priority:1" json:"test_name"`
	TestDescription string              `gorm:"size:1000" json:"test_description"`
	Status          TestExecutionStatus `gorm:"not null;index" json:"status"`

	FailureMessage string `json:"failure_message,omitempty" gorm:"not null;default:''"`

	StartTime time.Time `gorm:"not null;index:idx_test_name_time,priority:2" json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}
