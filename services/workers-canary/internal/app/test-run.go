package app

import (
	"time"

	"gorm.io/plugin/soft_delete"
)

type TestRunStatus string

const (
	TestRunPending  TestRunStatus = "pending"
	TestRunRunning  TestRunStatus = "running"
	TestRunFinished TestRunStatus = "finished"
	TestRunFailed   TestRunStatus = "failed"
	TestRunAborted  TestRunStatus = "aborted"
)

type TestRunType string

const (
	TestRunPullRequest TestRunType = "pull_request"
	TestRunManual      TestRunType = "manual"
	TestRunCron        TestRunType = "cron"
)

// TestRun represents a single canary lifecycle.
type TestRun struct {
	ID        string                `gorm:"primarykey;check:id_checker,char_length(id)=26" json:"id,omitzero" temporaljson:"id,omitzero,omitempty"`
	CreatedAt time.Time             `json:"created_at,omitzero" temporaljson:"created_at,omitzero,omitempty"`
	UpdatedAt time.Time             `json:"updated_at,omitzero" temporaljson:"updated_at,omitzero,omitempty"`
	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"index:idx_app_name,unique" temporaljson:"deleted_at,omitzero,omitempty"`

	MonoGithash string `json:"mono_githash,omitzero" gorm:"not null" temporaljson:"mono_githash,omitzero,omitempty"`
	PullRequest string `json:"pull_request,omitzero" gorm:"not null" temporaljson:"pull_request,omitzero,omitempty"`

	Status TestRunStatus `gorm:"not null;index" json:"status"`
	Type   TestRunType   `gorm:"not null;index" json:"type"`

	CronSchedule string `json:"cron_schedule,omitzero" gorm:"not null;default:''" temporaljson:"cron_schedule,omitzero,omitempty"`

	StartTime time.Time `gorm:"not null;index:idx_test_name_time,priority:2" json:"start_time"`
	EndTime   time.Time `json:"end_time"`

	CanaryGroups []TestGroup `faker:"-" json:"test_groups,omitempty" temporaljson:"test_groups,omitempty"`
}
