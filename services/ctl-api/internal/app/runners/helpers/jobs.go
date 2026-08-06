package helpers

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const (
	DefaultQueueTimeout     time.Duration = time.Hour * 24
	DefaultAvailableTimeout time.Duration = time.Minute * 1
	DefaultExecutionTimeout time.Duration = time.Minute * 5

	DefaultMaxExecutions int = 1
)

func (s *Helpers) getDefaultExecutionTimeout(typ app.RunnerJobType) time.Duration {
	timeouts := map[app.RunnerJobType]time.Duration{
		// build timeouts
		app.RunnerJobTypeDockerBuild:          time.Minute * 60,
		app.RunnerJobTypeContainerImageBuild:  time.Minute * 15,
		app.RunnerJobTypeHelmChartBuild:       time.Minute * 5,
		app.RunnerJobTypeTerraformModuleBuild: time.Minute * 15,
		app.RunnerJobTypePulumiBuild:          time.Minute * 5,

		// sync timeouts
		app.RunnerJobTypeOCISync:            time.Minute * 15,
		app.RunnerJobTypeFetchImageMetadata: time.Minute * 5,

		// deploy timeouts
		app.RunnerJobTypeTerraformDeploy:          time.Minute * 60,
		app.RunnerJobTypeHelmChartDeploy:          time.Minute * 30,
		app.RunnerJobTypeKubrenetesManifestDeploy: time.Minute * 15,
		app.RunnerJobTypePulumiDeploy:             time.Minute * 60,
		app.RunnerJobTypeJobDeploy:                time.Minute * 15,

		// sandbox timeouts
		app.RunnerJobTypeSandboxTerraform: time.Minute * 60,
		app.RunnerJobTypeSandboxPulumi:    time.Minute * 60,
		app.RunnerJobTypeRunnerTerraform:  time.Minute * 15,
		app.RunnerJobTypeRunnerHelm:       time.Minute * 5,
	}
	timeout, ok := timeouts[typ]
	if ok {
		return timeout
	}

	return DefaultExecutionTimeout
}

func (s *Helpers) getJob(ctx context.Context, jobID string) (*app.RunnerJob, error) {
	var runnerJob app.RunnerJob

	if res := s.db.WithContext(ctx).First(&runnerJob, "id = ?", jobID); res.Error != nil {
		return nil, fmt.Errorf("unable to get job: %w", res.Error)
	}

	return &runnerJob, nil
}

// CreateJobExecutionResultIfAbsent atomically preserves the first result written for an execution.
func CreateJobExecutionResultIfAbsent(ctx context.Context, db *gorm.DB, result *app.RunnerJobExecutionResult) (*app.RunnerJobExecutionResult, bool, error) {
	res := db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "deleted_at"},
				{Name: "runner_job_execution_id"},
			},
			DoNothing: true,
		}).
		Create(result)
	if res.Error != nil {
		return nil, false, fmt.Errorf("unable to create runner job execution result: %w", res.Error)
	}
	if res.RowsAffected == 1 {
		return result, true, nil
	}

	var existing app.RunnerJobExecutionResult
	if err := db.WithContext(ctx).
		Where(&app.RunnerJobExecutionResult{RunnerJobExecutionID: result.RunnerJobExecutionID}).
		First(&existing).Error; err != nil {
		return nil, false, fmt.Errorf("unable to get existing runner job execution result: %w", err)
	}
	return &existing, false, nil
}
