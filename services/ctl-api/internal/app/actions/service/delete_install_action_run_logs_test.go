package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	pkg_generics "github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type DeleteInstallActionRunLogsTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	L              *zap.Logger
	ActionsService *service
	Seeder         *testseed.Seeder
}

type DeleteInstallActionRunLogsTestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service DeleteInstallActionRunLogsTestService
	router  *gin.Engine
	ctx     context.Context
	testOrg *app.Org
	testAcc *app.Account
	testApp *app.App
}

func TestDeleteInstallActionRunLogsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(DeleteInstallActionRunLogsTestSuite))
}

func (s *DeleteInstallActionRunLogsTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)
	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T:               s.T(),
			CustomValidator: true,
		}),
		fx.Provide(New),
		fx.Populate(&s.service),
	)
	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.service.DB)
	s.SetCHDB(s.service.CHDB)
}

func (s *DeleteInstallActionRunLogsTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.ctx = context.Background()
	s.ctx, s.testAcc = s.service.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.service.Seeder.EnsureOrg(s.ctx, s.T())
	s.testApp = s.service.Seeder.CreateApp(s.ctx, s.T())
	s.setOrgAdmin(true)

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})
	require.NoError(s.T(), s.service.ActionsService.RegisterPublicRoutes(s.router))
}

func (s *DeleteInstallActionRunLogsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *DeleteInstallActionRunLogsTestSuite) setOrgAdmin(admin bool) {
	roleType := app.RoleTypeOrgAdmin
	if !admin {
		roleType = app.RoleTypeOrgBuilder
	}
	s.testAcc.Roles = []app.Role{{
		RoleType: roleType,
		OrgID:    pkg_generics.NewNullString(s.testOrg.ID),
	}}
}

type deleteLogsFixture struct {
	install   *app.Install
	run       *app.InstallActionWorkflowRun
	job       *app.RunnerJob
	logStream app.LogStream
	parentID  string
}

func (s *DeleteInstallActionRunLogsTestSuite) createFixture(status app.InstallActionWorkflowRunStatus) deleteLogsFixture {
	ctx := cctx.SetAccountContext(s.ctx, s.testAcc)
	ctx = cctx.SetOrgContext(ctx, s.testOrg)

	install := &app.Install{
		ID:    domains.NewInstallID(),
		Name:  fmt.Sprintf("test-install-%s", domains.NewInstallID()),
		AppID: s.testApp.ID,
	}
	require.NoError(s.T(), s.service.DB.WithContext(ctx).
		Omit("app_config_id", "app_sandbox_config_id", "app_runner_config_id").
		Create(install).Error)

	action := &app.ActionWorkflow{ID: domains.NewActionWorkflowID(), OrgID: s.testOrg.ID, AppID: s.testApp.ID, Name: "db-export-" + install.ID}
	require.NoError(s.T(), s.service.DB.WithContext(ctx).Create(action).Error)

	installAction := &app.InstallActionWorkflow{ID: domains.NewInstallActionWorkflowConfigID(), OrgID: s.testOrg.ID, InstallID: install.ID, ActionWorkflowID: action.ID}
	require.NoError(s.T(), s.service.DB.WithContext(ctx).Create(installAction).Error)

	run := &app.InstallActionWorkflowRun{
		ID:                      domains.NewInstallActionWorkflowRunID(),
		OrgID:                   s.testOrg.ID,
		InstallID:               install.ID,
		InstallActionWorkflowID: pkg_generics.NullString{NullString: sql.NullString{String: installAction.ID, Valid: true}},
		Status:                  status,
		StatusDescription:       string(status),
		StatusV2:                app.CompositeStatus{Status: app.Status(status)},
		TriggerType:             app.ActionWorkflowTriggerTypeManual,
	}
	require.NoError(s.T(), s.service.DB.WithContext(ctx).Create(run).Error)

	job := s.service.Seeder.CreateRunnerJob(ctx, s.T(), run.ID, "install_action_workflow_runs")

	parent := &app.LogStream{ID: domains.NewLogStreamID(), OrgID: s.testOrg.ID, OwnerID: install.ID, OwnerType: "workflows"}
	require.NoError(s.T(), s.service.DB.WithContext(ctx).Create(parent).Error)

	var logStream app.LogStream
	require.NoError(s.T(), s.service.DB.WithContext(ctx).Where(app.LogStream{ID: *job.LogStreamID}).First(&logStream).Error)
	require.NoError(s.T(), s.service.DB.WithContext(ctx).
		Model(&logStream).
		Update("parent_log_stream_id", parent.ID).Error)

	return deleteLogsFixture{install: install, run: run, job: job, logStream: logStream, parentID: parent.ID}
}

func (s *DeleteInstallActionRunLogsTestSuite) writeLogs(logStreamID, runnerJobID string, count int) {
	for i := 0; i < count; i++ {
		require.NoError(s.T(), s.service.CHDB.WithContext(s.ctx).Create(&app.OtelLogRecord{
			ID:          domains.NewOtelLogID(),
			OrgID:       s.testOrg.ID,
			LogStreamID: logStreamID,
			RunnerJobID: runnerJobID,
			Timestamp:   time.Now(),
			Body:        "AWS_SECRET_ACCESS_KEY=example",
		}).Error)
	}
}

func (s *DeleteInstallActionRunLogsTestSuite) writeSpans(runnerJobID string, count int) {
	for i := 0; i < count; i++ {
		require.NoError(s.T(), s.service.CHDB.WithContext(s.ctx).Create(&app.OtelTraceIngestion{
			ID:          domains.NewOtelTraceID(),
			RunnerJobID: runnerJobID,
			Timestamp:   time.Now(),
		}).Error)
	}
}

func (s *DeleteInstallActionRunLogsTestSuite) countLogs(logStreamID, runnerJobID string) int64 {
	var count int64
	q := s.service.CHDB.WithContext(s.ctx).Model(&app.OtelLogRecord{}).Where(app.OtelLogRecord{OrgID: s.testOrg.ID, LogStreamID: logStreamID})
	if runnerJobID != "" {
		q = q.Where(app.OtelLogRecord{RunnerJobID: runnerJobID})
	}
	require.NoError(s.T(), q.Count(&count).Error)
	return count
}

func (s *DeleteInstallActionRunLogsTestSuite) countSpans(runnerJobID string) int64 {
	var count int64
	require.NoError(s.T(), s.service.CHDB.WithContext(s.ctx).
		Model(&app.OtelTraceIngestion{}).
		Where(app.OtelTraceIngestion{RunnerJobID: runnerJobID}).
		Count(&count).Error)
	return count
}

func (s *DeleteInstallActionRunLogsTestSuite) deleteLogs(installID, runID string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/v1/installs/%s/actions/runs/%s/logs", installID, runID), nil)
	require.NoError(s.T(), err)
	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *DeleteInstallActionRunLogsTestSuite) TestDeletesRunLogsAndSpans() {
	f := s.createFixture(app.InstallActionRunStatusFinished)
	siblingJobID := domains.NewRunnerJobID()
	s.writeLogs(f.logStream.ID, f.job.ID, 5)
	s.writeLogs(f.parentID, f.job.ID, 3)
	s.writeLogs(f.parentID, siblingJobID, 2)
	s.writeSpans(f.job.ID, 2)
	s.writeSpans(siblingJobID, 1)

	rr := s.deleteLogs(f.install.ID, f.run.ID)
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

	var resp app.LogStream
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	require.NotNil(s.T(), resp.LogsDeletedAt)
	assert.Equal(s.T(), s.testAcc.ID, resp.LogsDeletedByID)

	assert.Zero(s.T(), s.countLogs(f.logStream.ID, ""))
	assert.Zero(s.T(), s.countLogs(f.parentID, f.job.ID))
	assert.EqualValues(s.T(), 2, s.countLogs(f.parentID, siblingJobID))
	assert.Zero(s.T(), s.countSpans(f.job.ID))
	assert.EqualValues(s.T(), 1, s.countSpans(siblingJobID))
}

func (s *DeleteInstallActionRunLogsTestSuite) TestRepeatDeleteKeepsOriginalMarker() {
	f := s.createFixture(app.InstallActionRunStatusFinished)
	s.writeLogs(f.logStream.ID, f.job.ID, 2)

	first := s.deleteLogs(f.install.ID, f.run.ID)
	require.Equal(s.T(), http.StatusOK, first.Code, first.Body.String())
	var firstResp app.LogStream
	require.NoError(s.T(), json.Unmarshal(first.Body.Bytes(), &firstResp))

	s.writeLogs(f.logStream.ID, f.job.ID, 1)

	second := s.deleteLogs(f.install.ID, f.run.ID)
	require.Equal(s.T(), http.StatusOK, second.Code, second.Body.String())
	var secondResp app.LogStream
	require.NoError(s.T(), json.Unmarshal(second.Body.Bytes(), &secondResp))

	assert.Zero(s.T(), s.countLogs(f.logStream.ID, ""))
	require.NotNil(s.T(), secondResp.LogsDeletedAt)
	assert.True(s.T(), firstResp.LogsDeletedAt.Equal(*secondResp.LogsDeletedAt))
}

func (s *DeleteInstallActionRunLogsTestSuite) TestRejectsNonAdmin() {
	f := s.createFixture(app.InstallActionRunStatusFinished)
	s.writeLogs(f.logStream.ID, f.job.ID, 2)
	s.setOrgAdmin(false)

	rr := s.deleteLogs(f.install.ID, f.run.ID)
	require.Equal(s.T(), http.StatusForbidden, rr.Code, rr.Body.String())
	assert.EqualValues(s.T(), 2, s.countLogs(f.logStream.ID, ""))
}

func (s *DeleteInstallActionRunLogsTestSuite) TestRejectsRunningRun() {
	for _, status := range []app.InstallActionWorkflowRunStatus{
		app.InstallActionRunStatusQueued,
		app.InstallActionRunStatusInProgress,
	} {
		s.Run(string(status), func() {
			f := s.createFixture(status)
			s.writeLogs(f.logStream.ID, f.job.ID, 2)

			rr := s.deleteLogs(f.install.ID, f.run.ID)
			require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())
			assert.EqualValues(s.T(), 2, s.countLogs(f.logStream.ID, ""))
		})
	}
}

func (s *DeleteInstallActionRunLogsTestSuite) TestDeletesCancelledRunWithOpenLogStream() {
	f := s.createFixture(app.InstallActionRunStatusCancelled)
	require.NoError(s.T(), s.service.DB.WithContext(s.ctx).
		Model(&f.logStream).
		Update("open", true).Error)
	s.writeLogs(f.logStream.ID, f.job.ID, 2)

	rr := s.deleteLogs(f.install.ID, f.run.ID)
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())
	assert.Zero(s.T(), s.countLogs(f.logStream.ID, ""))
}

func (s *DeleteInstallActionRunLogsTestSuite) TestNotFoundForOtherInstall() {
	f := s.createFixture(app.InstallActionRunStatusFinished)
	other := s.createFixture(app.InstallActionRunStatusFinished)
	s.writeLogs(f.logStream.ID, f.job.ID, 2)

	rr := s.deleteLogs(other.install.ID, f.run.ID)
	require.Equal(s.T(), http.StatusNotFound, rr.Code, rr.Body.String())
	assert.EqualValues(s.T(), 2, s.countLogs(f.logStream.ID, ""))
}
