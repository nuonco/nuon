package service

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/collector/pdata/ptrace/ptraceotlp"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type RunnerOtelWriteTracesTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type RunnerOtelWriteTracesTestSuite struct {
	tests.BaseDBTestSuite
	app            *fxtest.App
	service        RunnerOtelWriteTracesTestService
	router         *gin.Engine
	testOrg        *app.Org
	testAcc        *app.Account
	testRunner     *app.Runner
	testRunnerGrp  *app.RunnerGroup
	testLogStream  *app.LogStream
	testRunnerJob  *app.RunnerJob
	testRunnerExec *app.RunnerJobExecution
}

func TestRunnerOtelWriteTracesSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(RunnerOtelWriteTracesTestSuite))
}

func (s *RunnerOtelWriteTracesTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(s.T()),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.service.DB)
	s.SetCHDB(s.service.CHDB)
}

func (s *RunnerOtelWriteTracesTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create router with public routes
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})
	err := s.service.RunnersService.RegisterRunnerRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *RunnerOtelWriteTracesTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *RunnerOtelWriteTracesTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	s.testOrg = s.service.Seeder.CreateOrg(ctx, s.T())

	// Create runner group
	s.testRunnerGrp = &app.RunnerGroup{
		ID:        domains.NewRunnerGroupID(),
		OrgID:     s.testOrg.ID,
		OwnerID:   s.testOrg.ID,
		OwnerType: "org",
		Type:      app.RunnerGroupTypeOrg,
		Platform:  app.AppRunnerTypeAWSEKS,
	}
	err := s.service.DB.WithContext(ctx).Create(s.testRunnerGrp).Error
	require.NoError(s.T(), err)

	// Create runner
	s.testRunner = &app.Runner{
		ID:            domains.NewRunnerID(),
		OrgID:         s.testOrg.ID,
		RunnerGroupID: s.testRunnerGrp.ID,
		Name:          "test-runner",
		Status:        app.RunnerStatusActive,
	}
	err = s.service.DB.WithContext(ctx).Create(s.testRunner).Error
	require.NoError(s.T(), err)

	// Create log stream (needed as FK for runner job)
	s.testLogStream = &app.LogStream{
		ID:      domains.NewLogStreamID(),
		OrgID:   s.testOrg.ID,
		OwnerID: s.testOrg.ID,
		Open:    true,
	}
	err = s.service.DB.WithContext(ctx).Create(s.testLogStream).Error
	require.NoError(s.T(), err)

	// Create runner job
	s.testRunnerJob = &app.RunnerJob{
		ID:                domains.NewRunnerJobID(),
		OrgID:             s.testOrg.ID,
		RunnerID:          s.testRunner.ID,
		LogStreamID:       generics.ToPtr(s.testLogStream.ID),
		Status:            app.RunnerJobStatusQueued,
		StatusDescription: "test job",
		Group:             app.RunnerJobGroupBuild,
		Type:              app.RunnerJobTypeDockerBuild,
		Operation:         app.RunnerJobOperationTypeBuild,
		QueueTimeout:      5 * time.Minute,
		AvailableTimeout:  10 * time.Minute,
		ExecutionTimeout:  30 * time.Minute,
		OverallTimeout:    60 * time.Minute,
		MaxExecutions:     3,
	}
	err = s.service.DB.WithContext(ctx).Create(s.testRunnerJob).Error
	require.NoError(s.T(), err)

	// Create runner job execution
	s.testRunnerExec = &app.RunnerJobExecution{
		OrgID:       s.testOrg.ID,
		RunnerJobID: s.testRunnerJob.ID,
		Status:      app.RunnerJobExecutionStatusPending,
	}
	err = s.service.DB.WithContext(ctx).Create(s.testRunnerExec).Error
	require.NoError(s.T(), err)
}

func (s *RunnerOtelWriteTracesTestSuite) makeRequest(method, path string, body []byte) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, bytes.NewReader(body))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *RunnerOtelWriteTracesTestSuite) createTestTraceRequest(runnerID string, traceCount int) ptraceotlp.ExportRequest {
	req := ptraceotlp.NewExportRequest()
	traces := req.Traces()
	resourceSpans := traces.ResourceSpans()

	for i := 0; i < traceCount; i++ {
		rs := resourceSpans.AppendEmpty()

		// Set resource attributes
		resourceAttrs := rs.Resource().Attributes()
		resourceAttrs.PutStr("service.name", fmt.Sprintf("test-service-%d", i))
		resourceAttrs.PutStr("runner_group.id", s.testRunnerGrp.ID)

		// Add scope spans
		scopeSpan := rs.ScopeSpans().AppendEmpty()
		scopeSpan.Scope().SetName("test-scope")
		scopeSpan.Scope().SetVersion("1.0.0")

		// Add span
		span := scopeSpan.Spans().AppendEmpty()

		// Generate trace and span IDs
		traceID := pcommon.TraceID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, byte(i)})
		spanID := pcommon.SpanID([8]byte{1, 2, 3, 4, 5, 6, 7, byte(i)})

		span.SetTraceID(traceID)
		span.SetSpanID(spanID)
		span.SetName(fmt.Sprintf("test-span-%d", i))
		span.SetKind(ptrace.SpanKindServer)

		// Set timestamps
		startTime := time.Now().Add(-time.Duration(i+1) * time.Second)
		endTime := startTime.Add(100 * time.Millisecond)
		span.SetStartTimestamp(pcommon.NewTimestampFromTime(startTime))
		span.SetEndTimestamp(pcommon.NewTimestampFromTime(endTime))

		// Set span attributes with runner context
		spanAttrs := span.Attributes()
		spanAttrs.PutStr("runner_job.id", s.testRunnerJob.ID)
		spanAttrs.PutStr("runner_job_execution.id", s.testRunnerExec.ID)
		spanAttrs.PutStr("runner_job_execution_step.name", "test-step")
		spanAttrs.PutStr("test.key", fmt.Sprintf("test-value-%d", i))

		// Set status
		span.Status().SetCode(ptrace.StatusCodeOk)
		span.Status().SetMessage("success")

		// Add event
		event := span.Events().AppendEmpty()
		event.SetName(fmt.Sprintf("test-event-%d", i))
		event.SetTimestamp(pcommon.NewTimestampFromTime(startTime.Add(50 * time.Millisecond)))
		event.Attributes().PutStr("event.key", "event.value")
	}

	return req
}

// truncateTraces clears the otel_traces table between subtests.
func (s *RunnerOtelWriteTracesTestSuite) truncateTraces() {
	err := s.service.CHDB.Exec("TRUNCATE TABLE IF EXISTS otel_traces ON CLUSTER simple").Error
	require.NoError(s.T(), err)
}

// countTraces returns the number of traces for a given runner ID.
func (s *RunnerOtelWriteTracesTestSuite) countTraces(runnerID string) int64 {
	var count int64
	err := s.service.CHDB.
		Model(&app.OtelTraceIngestion{}).
		Where("runner_id = ?", runnerID).
		Count(&count).Error
	require.NoError(s.T(), err)
	return count
}

// traceRow holds scalar fields from otel_traces, avoiding complex CH types that GORM can't scan.
type traceRow struct {
	RunnerID               string `gorm:"column:runner_id"`
	RunnerGroupID          string `gorm:"column:runner_group_id"`
	RunnerJobID            string `gorm:"column:runner_job_id"`
	RunnerJobExecutionID   string `gorm:"column:runner_job_execution_id"`
	RunnerJobExecutionStep string `gorm:"column:runner_job_execution_step"`
	ServiceName            string `gorm:"column:service_name"`
	SpanName               string `gorm:"column:span_name"`
	SpanKind               string `gorm:"column:span_kind"`
	StatusCode             string `gorm:"column:status_code"`
	StatusMessage          string `gorm:"column:status_message"`
	ScopeName              string `gorm:"column:scope_name"`
	ScopeVersion           string `gorm:"column:scope_version"`
}

// queryTraceRows returns scalar trace fields for a runner, using raw SQL to avoid GORM scanning issues.
func (s *RunnerOtelWriteTracesTestSuite) queryTraceRows(runnerID string) []traceRow {
	var rows []traceRow
	err := s.service.CHDB.Raw(
		"SELECT runner_id, runner_group_id, runner_job_id, runner_job_execution_id, "+
			"runner_job_execution_step, service_name, span_name, span_kind, status_code, "+
			"status_message, scope_name, scope_version "+
			"FROM otel_traces WHERE runner_id = ? ORDER BY span_name ASC", runnerID).
		Scan(&rows).Error
	require.NoError(s.T(), err)
	return rows
}

func (s *RunnerOtelWriteTracesTestSuite) TestRunnerOtelWriteTraces() {
	testCases := []struct {
		name           string
		setupFunc      func() (string, []byte)
		expectedCode   int
		expectedTraces int
		validateFunc   func(string)
	}{
		{
			name: "successfully writes single trace",
			setupFunc: func() (string, []byte) {
				req := s.createTestTraceRequest(s.testRunner.ID, 1)
				jsonData, err := req.MarshalJSON()
				require.NoError(s.T(), err)
				return s.testRunner.ID, jsonData
			},
			expectedCode:   http.StatusOK,
			expectedTraces: 1,
			validateFunc: func(runnerID string) {
				rows := s.queryTraceRows(runnerID)
				require.Len(s.T(), rows, 1)
				row := rows[0]

				assert.Equal(s.T(), s.testRunner.ID, row.RunnerID)
				assert.Equal(s.T(), s.testRunnerGrp.ID, row.RunnerGroupID)
				assert.Equal(s.T(), s.testRunnerJob.ID, row.RunnerJobID)
				assert.Equal(s.T(), s.testRunnerExec.ID, row.RunnerJobExecutionID)
				assert.Equal(s.T(), "test-step", row.RunnerJobExecutionStep)
				assert.Equal(s.T(), "test-service-0", row.ServiceName)
				assert.Equal(s.T(), "test-span-0", row.SpanName)
				assert.Equal(s.T(), "Server", row.SpanKind)
				assert.Equal(s.T(), "Ok", row.StatusCode)
				assert.Equal(s.T(), "success", row.StatusMessage)
				assert.Equal(s.T(), "test-scope", row.ScopeName)
				assert.Equal(s.T(), "1.0.0", row.ScopeVersion)
			},
		},
		{
			name: "successfully writes multiple traces",
			setupFunc: func() (string, []byte) {
				req := s.createTestTraceRequest(s.testRunner.ID, 5)
				jsonData, err := req.MarshalJSON()
				require.NoError(s.T(), err)
				return s.testRunner.ID, jsonData
			},
			expectedCode:   http.StatusOK,
			expectedTraces: 5,
			validateFunc: func(runnerID string) {
				rows := s.queryTraceRows(runnerID)
				require.Len(s.T(), rows, 5)

				for i, row := range rows {
					assert.Equal(s.T(), s.testRunner.ID, row.RunnerID)
					assert.Equal(s.T(), s.testRunnerGrp.ID, row.RunnerGroupID)
					assert.Equal(s.T(), fmt.Sprintf("test-service-%d", i), row.ServiceName)
					assert.Equal(s.T(), fmt.Sprintf("test-span-%d", i), row.SpanName)
				}
			},
		},
		{
			name: "successfully writes trace with links",
			setupFunc: func() (string, []byte) {
				req := ptraceotlp.NewExportRequest()
				traces := req.Traces()
				rs := traces.ResourceSpans().AppendEmpty()

				resourceAttrs := rs.Resource().Attributes()
				resourceAttrs.PutStr("service.name", "test-service-with-links")
				resourceAttrs.PutStr("runner_group.id", s.testRunnerGrp.ID)

				scopeSpan := rs.ScopeSpans().AppendEmpty()
				scopeSpan.Scope().SetName("test-scope")

				span := scopeSpan.Spans().AppendEmpty()
				span.SetTraceID(pcommon.TraceID([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}))
				span.SetSpanID(pcommon.SpanID([8]byte{1, 2, 3, 4, 5, 6, 7, 8}))
				span.SetName("span-with-links")
				span.SetKind(ptrace.SpanKindClient)

				startTime := time.Now()
				span.SetStartTimestamp(pcommon.NewTimestampFromTime(startTime))
				span.SetEndTimestamp(pcommon.NewTimestampFromTime(startTime.Add(100 * time.Millisecond)))

				spanAttrs := span.Attributes()
				spanAttrs.PutStr("runner_job.id", s.testRunnerJob.ID)
				spanAttrs.PutStr("runner_job_execution.id", s.testRunnerExec.ID)
				spanAttrs.PutStr("runner_job_execution_step.name", "test-step")

				// Add links
				link1 := span.Links().AppendEmpty()
				link1.SetTraceID(pcommon.TraceID([16]byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}))
				link1.SetSpanID(pcommon.SpanID([8]byte{2, 2, 2, 2, 2, 2, 2, 2}))
				link1.Attributes().PutStr("link.key", "link.value")

				link2 := span.Links().AppendEmpty()
				link2.SetTraceID(pcommon.TraceID([16]byte{3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3}))
				link2.SetSpanID(pcommon.SpanID([8]byte{3, 3, 3, 3, 3, 3, 3, 3}))

				span.Status().SetCode(ptrace.StatusCodeOk)

				jsonData, err := req.MarshalJSON()
				require.NoError(s.T(), err)
				return s.testRunner.ID, jsonData
			},
			expectedCode:   http.StatusOK,
			expectedTraces: 1,
			validateFunc: func(runnerID string) {
				// Verify the trace was written with correct scalar fields
				rows := s.queryTraceRows(runnerID)
				require.Len(s.T(), rows, 1)
				assert.Equal(s.T(), "span-with-links", rows[0].SpanName)
				assert.Equal(s.T(), "Client", rows[0].SpanKind)

				// Verify links via raw SQL count
				var linkCount struct {
					Cnt uint64 `gorm:"column:cnt"`
				}
				err := s.service.CHDB.Raw(
					"SELECT length(`links.trace_id`) as cnt FROM otel_traces WHERE runner_id = ?", runnerID).
					Scan(&linkCount).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), uint64(2), linkCount.Cnt)
			},
		},
		{
			name: "empty request body returns error",
			setupFunc: func() (string, []byte) {
				return s.testRunner.ID, []byte("")
			},
			expectedCode:   http.StatusInternalServerError,
			expectedTraces: 0,
			validateFunc:   func(runnerID string) {},
		},
		{
			name: "invalid JSON returns error",
			setupFunc: func() (string, []byte) {
				return s.testRunner.ID, []byte("{invalid json}")
			},
			expectedCode:   http.StatusInternalServerError,
			expectedTraces: 0,
			validateFunc:   func(runnerID string) {},
		},
		{
			name: "malformed OTLP structure accepted as empty",
			setupFunc: func() (string, []byte) {
				return s.testRunner.ID, []byte(`{"not":"a valid otlp request"}`)
			},
			// UnmarshalJSON succeeds on valid JSON with unrecognized fields, producing zero traces.
			// The handler writes nothing and returns 200.
			expectedCode:   http.StatusOK,
			expectedTraces: 0,
			validateFunc: func(runnerID string) {
				count := s.countTraces(runnerID)
				assert.Equal(s.T(), int64(0), count)
			},
		},
		{
			name: "traces isolated by runner ID",
			setupFunc: func() (string, []byte) {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create second runner
				runner2 := &app.Runner{
					ID:            domains.NewRunnerID(),
					OrgID:         s.testOrg.ID,
					RunnerGroupID: s.testRunnerGrp.ID,
					Name:          "test-runner-2",
					Status:        app.RunnerStatusActive,
				}
				err := s.service.DB.WithContext(ctx).Create(runner2).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(runner2)
				})

				// Write traces for runner2
				req2 := s.createTestTraceRequest(runner2.ID, 3)
				jsonData2, err := req2.MarshalJSON()
				require.NoError(s.T(), err)

				path2 := fmt.Sprintf("/v1/runners/%s/traces", runner2.ID)
				rr2 := s.makeRequest("POST", path2, jsonData2)
				require.Equal(s.T(), http.StatusOK, rr2.Code)

				// Now write traces for original runner
				req := s.createTestTraceRequest(s.testRunner.ID, 2)
				jsonData, err := req.MarshalJSON()
				require.NoError(s.T(), err)
				return s.testRunner.ID, jsonData
			},
			expectedCode:   http.StatusOK,
			expectedTraces: 2,
			validateFunc: func(runnerID string) {
				rows := s.queryTraceRows(runnerID)
				require.Len(s.T(), rows, 2)
				for _, row := range rows {
					assert.Equal(s.T(), s.testRunner.ID, row.RunnerID)
				}
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Truncate CH traces between subtests to prevent data accumulation
			s.truncateTraces()

			runnerID, jsonData := tc.setupFunc()
			path := fmt.Sprintf("/v1/runners/%s/traces", runnerID)
			rr := s.makeRequest("POST", path, jsonData)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusOK && tc.expectedTraces > 0 {
				time.Sleep(200 * time.Millisecond)
				count := s.countTraces(runnerID)
				assert.Equal(s.T(), int64(tc.expectedTraces), count)
			}

			if tc.validateFunc != nil {
				tc.validateFunc(runnerID)
			}

			if tc.expectedCode != http.StatusOK {
				assert.Contains(s.T(), rr.Body.String(), "error")
			}
		})
	}
}
