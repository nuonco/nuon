package testseed

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// CreateRunnerJob persists a RunnerJob with the given polymorphic owner.
// Creates a LogStream, RunnerGroup, and Runner inline to satisfy FK constraints.
func (s *Seeder) CreateRunnerJob(ctx context.Context, t *testing.T, ownerID, ownerType string) *app.RunnerJob {
	orgID, _ := cctx.OrgIDFromContext(ctx)

	logStream := &app.LogStream{
		ID:        domains.NewLogStreamID(),
		OwnerID:   ownerID,
		OwnerType: ownerType,
		OrgID:     orgID,
		Open:      false,
	}
	require.NoError(t, s.db.WithContext(ctx).Create(logStream).Error)

	runnerGroup := &app.RunnerGroup{
		ID:        domains.NewRunnerGroupID(),
		OrgID:     orgID,
		OwnerID:   orgID,
		OwnerType: "org",
		Type:      app.RunnerGroupTypeOrg,
		Platform:  app.AppRunnerTypeAWSEKS,
	}
	require.NoError(t, s.db.WithContext(ctx).Create(runnerGroup).Error)

	runner := &app.Runner{
		ID:            domains.NewRunnerID(),
		OrgID:         orgID,
		Name:          "test-runner",
		DisplayName:   "Test Runner",
		Status:        app.RunnerStatusActive,
		RunnerGroupID: runnerGroup.ID,
	}
	require.NoError(t, s.db.WithContext(ctx).Create(runner).Error)

	job := &app.RunnerJob{
		LogStreamID:   &logStream.ID,
		RunnerID:      runner.ID,
		OwnerID:       ownerID,
		OwnerType:     ownerType,
		Status:        app.RunnerJobStatusFinished,
		Type:          app.RunnerJobTypeTerraformDeploy,
		Group:         app.RunnerJobGroupDeploy,
		Operation:     app.RunnerJobOperationTypeApplyPlan,
		MaxExecutions: 1,
	}
	require.NoError(t, s.db.WithContext(ctx).Create(job).Error)
	return job
}
