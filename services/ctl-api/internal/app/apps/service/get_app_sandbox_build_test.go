package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/compositeerrors"
)

func (s *AppsTestSuite) TestGetAppSandboxBuildIncludesLatestCompositeError() {
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, s.testAcc)
	ctx = cctx.SetOrgContext(ctx, s.testOrg)
	testApp := s.service.Seeder.CreateApp(ctx, s.T())
	appConfig := s.service.Seeder.CreateBareAppConfig(ctx, s.T(), testApp.ID)
	sandboxConfig := s.service.Seeder.CreateAppSandboxConfig(ctx, s.T(), testApp.ID, appConfig.ID)
	build := &app.AppSandboxBuild{
		AppID:              testApp.ID,
		AppConfigID:        appConfig.ID,
		AppSandboxConfigID: sandboxConfig.ID,
		Status:             app.AppSandboxBuildStatusError,
		StatusDescription:  "build failed",
	}
	require.NoError(s.T(), s.service.DB.WithContext(ctx).Create(build).Error)

	job := s.service.Seeder.CreateRunnerJob(ctx, s.T(), build.ID, "app_sandbox_builds")
	require.NoError(s.T(), s.service.DB.WithContext(ctx).
		Model(&app.RunnerJob{ID: job.ID}).
		Select("status").
		Updates(app.RunnerJob{Status: app.RunnerJobStatusFailed}).Error)
	execution := &app.RunnerJobExecution{
		RunnerJobID: job.ID,
		Status:      app.RunnerJobExecutionStatusFailed,
	}
	require.NoError(s.T(), s.service.DB.WithContext(ctx).Create(execution).Error)
	expected := &compositeerrors.CompositeErrorData{
		Version:  compositeerrors.SchemaVersion,
		Type:     "terraform.aws_permission",
		Severity: compositeerrors.SeverityError,
		Message:  "missing permission",
		Data:     json.RawMessage("null"),
	}
	require.NoError(s.T(), s.service.DB.WithContext(ctx).Create(&app.RunnerJobExecutionResult{
		RunnerJobExecutionID: execution.ID,
		CompositeError:       expected,
	}).Error)

	path := fmt.Sprintf("/v1/apps/%s/sandbox/builds/%s", testApp.ID, build.ID)
	rr := s.makeRequest(http.MethodGet, path)
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

	var response app.AppSandboxBuild
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &response))
	assert.Equal(s.T(), expected, response.CompositeError)
}
