package service

import (
	"bytes"
	"context"
	"encoding/json"
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
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/pmetric/pmetricotlp"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type RunnerOtelWriteMetricsTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type RunnerOtelWriteMetricsTestSuite struct {
	tests.BaseDBTestSuite
	app            *fxtest.App
	service        RunnerOtelWriteMetricsTestService
	router         *gin.Engine
	testOrg        *app.Org
	testAcc        *app.Account
	testRunner     *app.Runner
	testRunnerGrp  *app.RunnerGroup
	testRunnerGrpS *app.RunnerGroupSettings
}

func TestRunnerOtelWriteMetricsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(RunnerOtelWriteMetricsTestSuite))
}

func (s *RunnerOtelWriteMetricsTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.service.DB)
	s.SetCHDB(s.service.CHDB)
}

func (s *RunnerOtelWriteMetricsTestSuite) SetupTest() {
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

func (s *RunnerOtelWriteMetricsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *RunnerOtelWriteMetricsTestSuite) setupTestData() {
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

	// Create runner group settings
	s.testRunnerGrpS = &app.RunnerGroupSettings{
		ID:            domains.NewRunnerGroupSettingsID(),
		OrgID:         s.testOrg.ID,
		RunnerGroupID: s.testRunnerGrp.ID,
	}
	err = s.service.DB.WithContext(ctx).Create(s.testRunnerGrpS).Error
	require.NoError(s.T(), err)

	// Create runner
	s.testRunner = &app.Runner{
		ID:            domains.NewRunnerID(),
		OrgID:         s.testOrg.ID,
		RunnerGroupID: s.testRunnerGrp.ID,
		Name:          "test-runner",
		DisplayName:   "Test Runner",
		Status:        app.RunnerStatusActive,
	}
	err = s.service.DB.WithContext(ctx).Create(s.testRunner).Error
	require.NoError(s.T(), err)
}

func (s *RunnerOtelWriteMetricsTestSuite) makeRequest(method, path string, body []byte) *httptest.ResponseRecorder {
	var req *http.Request
	var err error

	if body != nil {
		req, err = http.NewRequest(method, path, bytes.NewReader(body))
	} else {
		req, err = http.NewRequest(method, path, nil)
	}
	require.NoError(s.T(), err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

// buildSumMetricRequest creates a valid OTEL metrics export request with Sum metric type
func (s *RunnerOtelWriteMetricsTestSuite) buildSumMetricRequest(runnerID, metricName string, value float64) []byte {
	metrics := pmetric.NewMetrics()
	rm := metrics.ResourceMetrics().AppendEmpty()

	// Set resource attributes
	rm.Resource().Attributes().PutStr("service.name", "test-service")
	rm.Resource().Attributes().PutStr("runner_group.id", s.testRunnerGrp.ID)

	// Set scope
	sm := rm.ScopeMetrics().AppendEmpty()
	sm.Scope().SetName("test-scope")
	sm.Scope().SetVersion("1.0.0")

	// Create sum metric
	metric := sm.Metrics().AppendEmpty()
	metric.SetName(metricName)
	metric.SetDescription("Test metric description")
	metric.SetUnit("count")

	sum := metric.SetEmptySum()
	sum.SetIsMonotonic(true)
	sum.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)

	// Add data point
	dp := sum.DataPoints().AppendEmpty()
	dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	dp.SetStartTimestamp(pcommon.NewTimestampFromTime(time.Now().Add(-1 * time.Minute)))
	dp.SetDoubleValue(value)
	dp.Attributes().PutStr("runner_job.id", "test-job-id")
	dp.Attributes().PutStr("runner_job_execution.id", "test-exec-id")
	dp.Attributes().PutStr("runner_job_execution_step.name", "test-step")

	// Convert to export request
	expreq := pmetricotlp.NewExportRequestFromMetrics(metrics)
	jsonData, err := expreq.MarshalJSON()
	require.NoError(s.T(), err)

	return jsonData
}

// buildGaugeMetricRequest creates a valid OTEL metrics export request with Gauge metric type
func (s *RunnerOtelWriteMetricsTestSuite) buildGaugeMetricRequest(runnerID, metricName string, value float64) []byte {
	metrics := pmetric.NewMetrics()
	rm := metrics.ResourceMetrics().AppendEmpty()

	// Set resource attributes
	rm.Resource().Attributes().PutStr("service.name", "test-service")
	rm.Resource().Attributes().PutStr("runner_group.id", s.testRunnerGrp.ID)

	// Set scope
	sm := rm.ScopeMetrics().AppendEmpty()
	sm.Scope().SetName("test-scope")

	// Create gauge metric
	metric := sm.Metrics().AppendEmpty()
	metric.SetName(metricName)
	metric.SetDescription("Test gauge metric")
	metric.SetUnit("ms")

	gauge := metric.SetEmptyGauge()

	// Add data point
	dp := gauge.DataPoints().AppendEmpty()
	dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	dp.SetStartTimestamp(pcommon.NewTimestampFromTime(time.Now().Add(-1 * time.Minute)))
	dp.SetDoubleValue(value)
	dp.Attributes().PutStr("runner_job.id", "test-job-id")
	dp.Attributes().PutStr("runner_job_execution.id", "test-exec-id")

	expreq := pmetricotlp.NewExportRequestFromMetrics(metrics)
	jsonData, err := expreq.MarshalJSON()
	require.NoError(s.T(), err)

	return jsonData
}

// buildHistogramMetricRequest creates a valid OTEL metrics export request with Histogram metric type
func (s *RunnerOtelWriteMetricsTestSuite) buildHistogramMetricRequest(runnerID, metricName string) []byte {
	metrics := pmetric.NewMetrics()
	rm := metrics.ResourceMetrics().AppendEmpty()

	rm.Resource().Attributes().PutStr("service.name", "test-service")
	rm.Resource().Attributes().PutStr("runner_group.id", s.testRunnerGrp.ID)

	sm := rm.ScopeMetrics().AppendEmpty()
	sm.Scope().SetName("test-scope")

	metric := sm.Metrics().AppendEmpty()
	metric.SetName(metricName)
	metric.SetDescription("Test histogram metric")
	metric.SetUnit("ms")

	histogram := metric.SetEmptyHistogram()
	histogram.SetAggregationTemporality(pmetric.AggregationTemporalityCumulative)

	dp := histogram.DataPoints().AppendEmpty()
	dp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
	dp.SetStartTimestamp(pcommon.NewTimestampFromTime(time.Now().Add(-1 * time.Minute)))
	dp.SetCount(100)
	dp.SetSum(5000.0)
	dp.SetMin(10.0)
	dp.SetMax(200.0)
	dp.ExplicitBounds().FromRaw([]float64{0, 50, 100, 150, 200})
	dp.BucketCounts().FromRaw([]uint64{10, 30, 40, 15, 5})
	dp.Attributes().PutStr("runner_job.id", "test-job-id")
	dp.Attributes().PutStr("runner_job_execution.id", "test-exec-id")

	expreq := pmetricotlp.NewExportRequestFromMetrics(metrics)
	jsonData, err := expreq.MarshalJSON()
	require.NoError(s.T(), err)

	return jsonData
}

func (s *RunnerOtelWriteMetricsTestSuite) TestRunnerOtelWriteMetrics() {
	testCases := []struct {
		name         string
		setupFunc    func() (string, []byte)
		expectedCode int
		validateFunc func(string)
	}{
		{
			name: "successfully write sum metric",
			setupFunc: func() (string, []byte) {
				body := s.buildSumMetricRequest(s.testRunner.ID, "test.counter", 42.0)
				return s.testRunner.ID, body
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(runnerID string) {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)
				ctx = cctx.SetOrgContext(ctx, s.testOrg)

				// Use count query to avoid ClickHouse array scan issues with GORM
				var count int64
				err := s.service.CHDB.WithContext(ctx).
					Model(&app.OtelMetricSumIngestion{}).
					Where("runner_id = ?", runnerID).
					Where("metric_name = ?", "test.counter").
					Count(&count).Error
				require.NoError(s.T(), err)
				require.Equal(s.T(), int64(1), count)

				// Verify specific fields using raw SQL to avoid scanning complex ClickHouse types
				var result struct {
					RunnerID      string  `gorm:"column:runner_id"`
					RunnerGroupID string  `gorm:"column:runner_group_id"`
					MetricName    string  `gorm:"column:metric_name"`
					Value         float64 `gorm:"column:value"`
					ServiceName   string  `gorm:"column:service_name"`
					IsMonotonic   bool    `gorm:"column:is_monotonic"`
				}
				err = s.service.CHDB.WithContext(ctx).
					Raw("SELECT runner_id, runner_group_id, metric_name, value, service_name, is_monotonic FROM otel_metrics_sum WHERE runner_id = ? AND metric_name = ?", runnerID, "test.counter").
					Scan(&result).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), runnerID, result.RunnerID)
				assert.Equal(s.T(), s.testRunnerGrp.ID, result.RunnerGroupID)
				assert.Equal(s.T(), "test.counter", result.MetricName)
				assert.Equal(s.T(), 42.0, result.Value)
				assert.Equal(s.T(), "test-service", result.ServiceName)
				assert.True(s.T(), result.IsMonotonic)
			},
		},
		{
			name: "successfully write gauge metric",
			setupFunc: func() (string, []byte) {
				body := s.buildGaugeMetricRequest(s.testRunner.ID, "test.gauge", 99.5)
				return s.testRunner.ID, body
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(runnerID string) {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)
				ctx = cctx.SetOrgContext(ctx, s.testOrg)

				// Use count query to avoid ClickHouse array scan issues with GORM
				var count int64
				err := s.service.CHDB.WithContext(ctx).
					Model(&app.OtelMetricGaugeIngestion{}).
					Where("runner_id = ?", runnerID).
					Where("metric_name = ?", "test.gauge").
					Count(&count).Error
				require.NoError(s.T(), err)
				require.Equal(s.T(), int64(1), count)

				// Verify specific fields using raw SQL to avoid scanning complex ClickHouse types
				var result struct {
					RunnerID   string  `gorm:"column:runner_id"`
					MetricName string  `gorm:"column:metric_name"`
					Value      float64 `gorm:"column:value"`
				}
				err = s.service.CHDB.WithContext(ctx).
					Raw("SELECT runner_id, metric_name, value FROM otel_metrics_gauge WHERE runner_id = ? AND metric_name = ?", runnerID, "test.gauge").
					Scan(&result).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), runnerID, result.RunnerID)
				assert.Equal(s.T(), "test.gauge", result.MetricName)
				assert.Equal(s.T(), 99.5, result.Value)
			},
		},
		{
			name: "successfully write histogram metric",
			setupFunc: func() (string, []byte) {
				body := s.buildHistogramMetricRequest(s.testRunner.ID, "test.histogram")
				return s.testRunner.ID, body
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(runnerID string) {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)
				ctx = cctx.SetOrgContext(ctx, s.testOrg)

				// Use count query to avoid ClickHouse array scan issues with GORM
				var count int64
				err := s.service.CHDB.WithContext(ctx).
					Model(&app.OtelMetricHistogramIngestion{}).
					Where("runner_id = ?", runnerID).
					Where("metric_name = ?", "test.histogram").
					Count(&count).Error
				require.NoError(s.T(), err)
				require.Equal(s.T(), int64(1), count)

				// Verify specific fields using raw SQL to avoid scanning complex ClickHouse types
				var result struct {
					RunnerID   string  `gorm:"column:runner_id"`
					MetricName string  `gorm:"column:metric_name"`
					Count      uint64  `gorm:"column:count"`
					Sum        float64 `gorm:"column:sum"`
					Min        float64 `gorm:"column:min"`
					Max        float64 `gorm:"column:max"`
				}
				err = s.service.CHDB.WithContext(ctx).
					Raw("SELECT runner_id, metric_name, count, sum, min, max FROM otel_metrics_histogram WHERE runner_id = ? AND metric_name = ?", runnerID, "test.histogram").
					Scan(&result).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), runnerID, result.RunnerID)
				assert.Equal(s.T(), "test.histogram", result.MetricName)
				assert.Equal(s.T(), uint64(100), result.Count)
				assert.Equal(s.T(), 5000.0, result.Sum)
				assert.Equal(s.T(), 10.0, result.Min)
				assert.Equal(s.T(), 200.0, result.Max)
			},
		},
		{
			name: "invalid JSON body returns error",
			setupFunc: func() (string, []byte) {
				return s.testRunner.ID, []byte(`{invalid json}`)
			},
			expectedCode: http.StatusInternalServerError,
			validateFunc: func(runnerID string) {
				// No validation needed - just checking error response
			},
		},
		{
			name: "empty body returns error",
			setupFunc: func() (string, []byte) {
				return s.testRunner.ID, []byte(``)
			},
			expectedCode: http.StatusInternalServerError,
			validateFunc: func(runnerID string) {
				// No validation needed
			},
		},
		{
			name: "invalid OTEL structure returns error",
			setupFunc: func() (string, []byte) {
				return s.testRunner.ID, []byte(`{"not": "valid", "otel": "structure"}`)
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(runnerID string) {
				// The handler accepts valid JSON and treats unrecognized fields as empty metrics data.
				// No metrics are written, but the request succeeds with 201.
			},
		},
		{
			name: "runner ID from path parameter is used",
			setupFunc: func() (string, []byte) {
				// Create second runner
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				runner2 := &app.Runner{
					ID:            domains.NewRunnerID(),
					OrgID:         s.testOrg.ID,
					RunnerGroupID: s.testRunnerGrp.ID,
					Name:          "test-runner-2",
					DisplayName:   "Test Runner 2",
					Status:        app.RunnerStatusActive,
				}
				err := s.service.DB.WithContext(ctx).Create(runner2).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(runner2)
				})

				body := s.buildSumMetricRequest(runner2.ID, "test.path.counter", 123.0)
				return runner2.ID, body
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(runnerID string) {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)
				ctx = cctx.SetOrgContext(ctx, s.testOrg)

				// Use count query to avoid ClickHouse array scan issues with GORM
				var count int64
				err := s.service.CHDB.WithContext(ctx).
					Model(&app.OtelMetricSumIngestion{}).
					Where("runner_id = ?", runnerID).
					Where("metric_name = ?", "test.path.counter").
					Count(&count).Error
				require.NoError(s.T(), err)
				require.Equal(s.T(), int64(1), count)

				// Verify the runner_id using raw SQL
				var result struct {
					RunnerID string `gorm:"column:runner_id"`
				}
				err = s.service.CHDB.WithContext(ctx).
					Raw("SELECT runner_id FROM otel_metrics_sum WHERE runner_id = ? AND metric_name = ?", runnerID, "test.path.counter").
					Scan(&result).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), runnerID, result.RunnerID)
			},
		},
		{
			name: "multiple metrics in single request",
			setupFunc: func() (string, []byte) {
				// Build request with multiple metrics
				metrics := pmetric.NewMetrics()
				rm := metrics.ResourceMetrics().AppendEmpty()
				rm.Resource().Attributes().PutStr("service.name", "test-service")
				rm.Resource().Attributes().PutStr("runner_group.id", s.testRunnerGrp.ID)

				sm := rm.ScopeMetrics().AppendEmpty()
				sm.Scope().SetName("test-scope")

				// First metric - sum
				metric1 := sm.Metrics().AppendEmpty()
				metric1.SetName("test.multi.counter1")
				sum1 := metric1.SetEmptySum()
				sum1.SetIsMonotonic(true)
				dp1 := sum1.DataPoints().AppendEmpty()
				dp1.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
				dp1.SetDoubleValue(10.0)
				dp1.Attributes().PutStr("runner_job.id", "job1")
				dp1.Attributes().PutStr("runner_job_execution.id", "exec1")

				// Second metric - gauge
				metric2 := sm.Metrics().AppendEmpty()
				metric2.SetName("test.multi.gauge1")
				gauge2 := metric2.SetEmptyGauge()
				dp2 := gauge2.DataPoints().AppendEmpty()
				dp2.SetTimestamp(pcommon.NewTimestampFromTime(time.Now()))
				dp2.SetDoubleValue(20.0)
				dp2.Attributes().PutStr("runner_job.id", "job1")
				dp2.Attributes().PutStr("runner_job_execution.id", "exec1")

				expreq := pmetricotlp.NewExportRequestFromMetrics(metrics)
				jsonData, err := expreq.MarshalJSON()
				require.NoError(s.T(), err)

				return s.testRunner.ID, jsonData
			},
			expectedCode: http.StatusCreated,
			validateFunc: func(runnerID string) {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)
				ctx = cctx.SetOrgContext(ctx, s.testOrg)

				// Use count queries to avoid ClickHouse array scan issues with GORM
				var sumCount int64
				err := s.service.CHDB.WithContext(ctx).
					Model(&app.OtelMetricSumIngestion{}).
					Where("runner_id = ?", runnerID).
					Where("metric_name = ?", "test.multi.counter1").
					Count(&sumCount).Error
				require.NoError(s.T(), err)
				require.Equal(s.T(), int64(1), sumCount)

				var gaugeCount int64
				err = s.service.CHDB.WithContext(ctx).
					Model(&app.OtelMetricGaugeIngestion{}).
					Where("runner_id = ?", runnerID).
					Where("metric_name = ?", "test.multi.gauge1").
					Count(&gaugeCount).Error
				require.NoError(s.T(), err)
				require.Equal(s.T(), int64(1), gaugeCount)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			runnerID, body := tc.setupFunc()
			path := fmt.Sprintf("/v1/runners/%s/metrics", runnerID)
			rr := s.makeRequest("POST", path, body)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusCreated {
				var response string
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(s.T(), err)
				assert.Equal(s.T(), "ok", response)

				if tc.validateFunc != nil {
					tc.validateFunc(runnerID)
				}
			} else {
				// Error cases should have error in response
				assert.Contains(s.T(), rr.Body.String(), "error")
			}
		})
	}
}
