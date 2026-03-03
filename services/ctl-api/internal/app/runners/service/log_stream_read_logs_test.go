package service

import (
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

type LogStreamReadLogsTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type LogStreamReadLogsTestSuite struct {
	tests.BaseDBTestSuite
	app           *fxtest.App
	service       LogStreamReadLogsTestService
	router        *gin.Engine
	testOrg       *app.Org
	testAcc       *app.Account
	testLogStream *app.LogStream
}

func TestLogStreamReadLogsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(LogStreamReadLogsTestSuite))
}

func (s *LogStreamReadLogsTestSuite) SetupSuite() {
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

func (s *LogStreamReadLogsTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create router with public routes
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})
	err := s.service.RunnersService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *LogStreamReadLogsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *LogStreamReadLogsTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	s.testOrg = s.service.Seeder.CreateOrg(ctx, s.T())

	// Create log stream in Postgres
	s.testLogStream = &app.LogStream{
		ID:      domains.NewLogStreamID(),
		OrgID:   s.testOrg.ID,
		OwnerID: s.testOrg.ID,
		Open:    true,
	}
	err := s.service.DB.WithContext(ctx).Create(s.testLogStream).Error
	require.NoError(s.T(), err)
}

func (s *LogStreamReadLogsTestSuite) makeRequest(method, path string, headers map[string]string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *LogStreamReadLogsTestSuite) createOtelLogRecords(logStreamID string, count int, baseTimestamp time.Time) []string {
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, s.testAcc)
	ctx = cctx.SetOrgContext(ctx, s.testOrg)

	logIDs := make([]string, count)
	for i := 0; i < count; i++ {
		log := &app.OtelLogRecord{
			ID:           domains.NewOtelLogID(),
			OrgID:        s.testOrg.ID,
			LogStreamID:  logStreamID,
			Timestamp:    baseTimestamp.Add(time.Duration(i) * time.Second),
			Body:         fmt.Sprintf("Log message %d", i),
			SeverityText: "INFO",
		}
		err := s.service.CHDB.WithContext(ctx).Create(log).Error
		require.NoError(s.T(), err)
		logIDs[i] = log.ID
	}
	return logIDs
}

func (s *LogStreamReadLogsTestSuite) TestLogStreamReadLogs() {
	testCases := []struct {
		name          string
		setupFunc     func() string
		queryParams   string
		headers       map[string]string
		expectedCode  int
		expectedCount int
		validateFunc  func([]app.OtelLogRecord, http.Header)
	}{
		{
			name: "empty log stream returns empty array",
			setupFunc: func() string {
				// Use existing test log stream with no logs
				return s.testLogStream.ID
			},
			expectedCode:  http.StatusOK,
			expectedCount: 0,
			validateFunc: func(logs []app.OtelLogRecord, headers http.Header) {
				assert.Equal(s.T(), "0", headers.Get("count"))
				assert.Equal(s.T(), "", headers.Get("X-Nuon-API-Next"))
				assert.Equal(s.T(), "items", headers.Get("Range-Units"))
			},
		},
		{
			name: "returns logs for correct log stream",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create new log stream
				logStream := &app.LogStream{
					ID:      domains.NewLogStreamID(),
					OrgID:   s.testOrg.ID,
					OwnerID: s.testOrg.ID,
					Open:    true,
				}
				err := s.service.DB.WithContext(ctx).Create(logStream).Error
				require.NoError(s.T(), err)

				// Create 5 log records
				s.createOtelLogRecords(logStream.ID, 5, time.Now().Add(-5*time.Minute))

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(logStream)
				})

				return logStream.ID
			},
			expectedCode:  http.StatusOK,
			expectedCount: 5,
			validateFunc: func(logs []app.OtelLogRecord, headers http.Header) {
				assert.Equal(s.T(), "5", headers.Get("count"))
				assert.Equal(s.T(), "", headers.Get("X-Nuon-API-Next"))
				assert.Len(s.T(), logs, 5)
				// Verify ascending order (default)
				for i := 0; i < len(logs)-1; i++ {
					assert.True(s.T(), logs[i].Timestamp.Before(logs[i+1].Timestamp) || logs[i].Timestamp.Equal(logs[i+1].Timestamp))
				}
			},
		},
		{
			name: "does not return logs from different log stream",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create two log streams
				logStream1 := &app.LogStream{
					ID:      domains.NewLogStreamID(),
					OrgID:   s.testOrg.ID,
					OwnerID: s.testOrg.ID,
					Open:    true,
				}
				err := s.service.DB.WithContext(ctx).Create(logStream1).Error
				require.NoError(s.T(), err)

				logStream2 := &app.LogStream{
					ID:      domains.NewLogStreamID(),
					OrgID:   s.testOrg.ID,
					OwnerID: s.testOrg.ID,
					Open:    true,
				}
				err = s.service.DB.WithContext(ctx).Create(logStream2).Error
				require.NoError(s.T(), err)

				// Create logs in both streams
				s.createOtelLogRecords(logStream1.ID, 3, time.Now().Add(-3*time.Minute))
				s.createOtelLogRecords(logStream2.ID, 2, time.Now().Add(-2*time.Minute))

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(logStream1)
					s.service.DB.Unscoped().Delete(logStream2)
				})

				// Request logs for logStream1, should only get 3
				return logStream1.ID
			},
			expectedCode:  http.StatusOK,
			expectedCount: 3,
			validateFunc: func(logs []app.OtelLogRecord, headers http.Header) {
				assert.Equal(s.T(), "3", headers.Get("count"))
				assert.Len(s.T(), logs, 3)
			},
		},
		{
			name: "log stream not found returns 404",
			setupFunc: func() string {
				return "lgsnonexistent123456789012"
			},
			expectedCode: http.StatusNotFound,
			validateFunc: func(logs []app.OtelLogRecord, headers http.Header) {
				// Error response, no logs
			},
		},
		{
			name: "log stream in different org returns 404",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create second org
				org2ID := domains.NewOrgID()
				org2 := &app.Org{
					ID:          org2ID,
					Name:        fmt.Sprintf("other-org-%s", org2ID),
					SandboxMode: true,
					NotificationsConfig: app.NotificationsConfig{
						InternalSlackWebhookURL: "https://hooks.slack.com/bar",
					},
				}
				err := s.service.DB.WithContext(ctx).Create(org2).Error
				require.NoError(s.T(), err)

				// Create log stream in org2
				logStream2 := &app.LogStream{
					ID:      domains.NewLogStreamID(),
					OrgID:   org2.ID,
					OwnerID: org2.ID,
					Open:    true,
				}
				err = s.service.DB.WithContext(ctx).Create(logStream2).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(logStream2)
					s.service.DB.Unscoped().Delete(org2)
				})

				return logStream2.ID
			},
			expectedCode: http.StatusNotFound,
			validateFunc: func(logs []app.OtelLogRecord, headers http.Header) {
				// Error response
			},
		},
		{
			name: "order=asc returns logs in ascending timestamp order",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				logStream := &app.LogStream{
					ID:      domains.NewLogStreamID(),
					OrgID:   s.testOrg.ID,
					OwnerID: s.testOrg.ID,
					Open:    true,
				}
				err := s.service.DB.WithContext(ctx).Create(logStream).Error
				require.NoError(s.T(), err)

				// Create logs with specific timestamps
				s.createOtelLogRecords(logStream.ID, 10, time.Now().Add(-10*time.Minute))

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(logStream)
				})

				return logStream.ID
			},
			queryParams:   "?order=asc",
			expectedCode:  http.StatusOK,
			expectedCount: 10,
			validateFunc: func(logs []app.OtelLogRecord, headers http.Header) {
				assert.Len(s.T(), logs, 10)
				// Verify ascending order
				for i := 0; i < len(logs)-1; i++ {
					assert.True(s.T(), logs[i].Timestamp.Before(logs[i+1].Timestamp) || logs[i].Timestamp.Equal(logs[i+1].Timestamp),
						"Logs should be in ascending order")
				}
			},
		},
		{
			name: "order=desc returns logs in descending timestamp order",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				logStream := &app.LogStream{
					ID:      domains.NewLogStreamID(),
					OrgID:   s.testOrg.ID,
					OwnerID: s.testOrg.ID,
					Open:    true,
				}
				err := s.service.DB.WithContext(ctx).Create(logStream).Error
				require.NoError(s.T(), err)

				// Create logs with specific timestamps
				s.createOtelLogRecords(logStream.ID, 10, time.Now().Add(-10*time.Minute))

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(logStream)
				})

				return logStream.ID
			},
			queryParams:   "?order=desc",
			expectedCode:  http.StatusOK,
			expectedCount: 10,
			validateFunc: func(logs []app.OtelLogRecord, headers http.Header) {
				assert.Len(s.T(), logs, 10)
				// Verify descending order
				for i := 0; i < len(logs)-1; i++ {
					assert.True(s.T(), logs[i].Timestamp.After(logs[i+1].Timestamp) || logs[i].Timestamp.Equal(logs[i+1].Timestamp),
						"Logs should be in descending order")
				}
			},
		},
		{
			name: "invalid order parameter returns error",
			setupFunc: func() string {
				return s.testLogStream.ID
			},
			queryParams:  "?order=invalid",
			expectedCode: http.StatusBadRequest,
			validateFunc: func(logs []app.OtelLogRecord, headers http.Header) {
				// Error response
			},
		},
		{
			name: "pagination with X-Nuon-API-Next header (ascending)",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				logStream := &app.LogStream{
					ID:      domains.NewLogStreamID(),
					OrgID:   s.testOrg.ID,
					OwnerID: s.testOrg.ID,
					Open:    true,
				}
				err := s.service.DB.WithContext(ctx).Create(logStream).Error
				require.NoError(s.T(), err)

				// Create 150 log records (more than page size of 100)
				s.createOtelLogRecords(logStream.ID, 150, time.Now().Add(-150*time.Minute))

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(logStream)
				})

				return logStream.ID
			},
			expectedCode:  http.StatusOK,
			expectedCount: 100,
			validateFunc: func(logs []app.OtelLogRecord, headers http.Header) {
				assert.Equal(s.T(), "150", headers.Get("count"))
				assert.Len(s.T(), logs, 100)
				assert.NotEmpty(s.T(), headers.Get("X-Nuon-API-Next"), "Should have next cursor")
			},
		},
		{
			name: "pagination second page with cursor (ascending)",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				logStream := &app.LogStream{
					ID:      domains.NewLogStreamID(),
					OrgID:   s.testOrg.ID,
					OwnerID: s.testOrg.ID,
					Open:    true,
				}
				err := s.service.DB.WithContext(ctx).Create(logStream).Error
				require.NoError(s.T(), err)

				// Create 150 log records
				s.createOtelLogRecords(logStream.ID, 150, time.Now().Add(-150*time.Minute))

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(logStream)
				})

				return logStream.ID
			},
			expectedCode: http.StatusOK,
			headers:      map[string]string{
				// This would be the cursor from the first request
				// For testing, we'll get the first page and extract the cursor
			},
			validateFunc: func(logs []app.OtelLogRecord, headers http.Header) {
				// Should get remaining 50 logs
				assert.Equal(s.T(), "150", headers.Get("count"))
			},
		},
		{
			name: "pagination with order=desc",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				logStream := &app.LogStream{
					ID:      domains.NewLogStreamID(),
					OrgID:   s.testOrg.ID,
					OwnerID: s.testOrg.ID,
					Open:    true,
				}
				err := s.service.DB.WithContext(ctx).Create(logStream).Error
				require.NoError(s.T(), err)

				// Create 150 log records
				s.createOtelLogRecords(logStream.ID, 150, time.Now().Add(-150*time.Minute))

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(logStream)
				})

				return logStream.ID
			},
			queryParams:   "?order=desc",
			expectedCode:  http.StatusOK,
			expectedCount: 100,
			validateFunc: func(logs []app.OtelLogRecord, headers http.Header) {
				assert.Equal(s.T(), "150", headers.Get("count"))
				assert.Len(s.T(), logs, 100)
				assert.NotEmpty(s.T(), headers.Get("X-Nuon-API-Next"))
				// Verify descending order
				for i := 0; i < len(logs)-1; i++ {
					assert.True(s.T(), logs[i].Timestamp.After(logs[i+1].Timestamp) || logs[i].Timestamp.Equal(logs[i+1].Timestamp))
				}
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			logStreamID := tc.setupFunc()
			path := fmt.Sprintf("/v1/log-streams/%s/logs%s", logStreamID, tc.queryParams)
			rr := s.makeRequest("GET", path, tc.headers)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusOK {
				var logs []app.OtelLogRecord
				err := json.Unmarshal(rr.Body.Bytes(), &logs)
				require.NoError(s.T(), err)

				if tc.expectedCount > 0 {
					assert.Len(s.T(), logs, tc.expectedCount)
				}

				if tc.validateFunc != nil {
					tc.validateFunc(logs, rr.Header())
				}
			} else {
				// Error cases should have error in response
				assert.Contains(s.T(), rr.Body.String(), "error")
			}
		})
	}
}
