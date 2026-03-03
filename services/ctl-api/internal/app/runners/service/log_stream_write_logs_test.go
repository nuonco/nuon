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
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/plog/plogotlp"
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

type LogStreamWriteLogsTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type LogStreamWriteLogsTestSuite struct {
	tests.BaseDBTestSuite
	app           *fxtest.App
	service       LogStreamWriteLogsTestService
	router        *gin.Engine
	testOrg       *app.Org
	testAcc       *app.Account
	testLogStream *app.LogStream
}

func TestLogStreamWriteLogsSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(LogStreamWriteLogsTestSuite))
}

func (s *LogStreamWriteLogsTestSuite) SetupSuite() {
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

func (s *LogStreamWriteLogsTestSuite) SetupTest() {
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

func (s *LogStreamWriteLogsTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *LogStreamWriteLogsTestSuite) setupTestData() {
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

func (s *LogStreamWriteLogsTestSuite) makeRequest(method, path string, body []byte) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		reqBody = bytes.NewBuffer(body)
	} else {
		reqBody = bytes.NewBuffer([]byte{})
	}

	req, err := http.NewRequest(method, path, reqBody)
	require.NoError(s.T(), err)

	// Set proper content type for protobuf binary
	if body != nil {
		req.Header.Set("Content-Type", "application/x-protobuf")
	}

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

// buildOTLPLogExportRequest creates a valid OTEL protobuf export request
func (s *LogStreamWriteLogsTestSuite) buildOTLPLogExportRequest(logMessages []string) []byte {
	exportReq := plogotlp.NewExportRequest()
	logs := exportReq.Logs()

	// Add resource logs
	resourceLogs := logs.ResourceLogs().AppendEmpty()

	// Set resource attributes
	resourceAttrs := resourceLogs.Resource().Attributes()
	resourceAttrs.PutStr("service.name", "test-service")
	resourceAttrs.PutStr("runner_group.id", "rgrp123456789")

	// Add scope logs
	scopeLogs := resourceLogs.ScopeLogs().AppendEmpty()
	scopeLogs.Scope().SetName("test-scope")
	scopeLogs.Scope().SetVersion("1.0.0")

	// Add log records
	for i, msg := range logMessages {
		logRecord := scopeLogs.LogRecords().AppendEmpty()
		logRecord.SetTimestamp(pcommon.NewTimestampFromTime(time.Now().Add(time.Duration(i) * time.Second)))
		logRecord.SetSeverityNumber(plog.SeverityNumberInfo)
		logRecord.SetSeverityText("INFO")
		logRecord.Body().SetStr(msg)

		// Add log attributes
		logAttrs := logRecord.Attributes()
		logAttrs.PutStr("runner.id", "rnr123456789")
		logAttrs.PutStr("runner_job.id", "rjb123456789")
	}

	// Convert to export request and marshal to protobuf binary
	protoData, err := exportReq.MarshalProto()
	require.NoError(s.T(), err)

	return protoData
}

// buildInvalidOTLPRequest creates an invalid protobuf payload
func (s *LogStreamWriteLogsTestSuite) buildInvalidOTLPRequest() []byte {
	return []byte("invalid protobuf data")
}

// countLogsInCH counts logs for a specific log stream in ClickHouse
func (s *LogStreamWriteLogsTestSuite) countLogsInCH(logStreamID string) int {
	var count int64
	err := s.service.CHDB.Model(&app.OtelLogRecord{}).
		Where("log_stream_id = ?", logStreamID).
		Count(&count).Error
	require.NoError(s.T(), err)
	return int(count)
}

// getLogsFromCH retrieves logs for a specific log stream from ClickHouse
func (s *LogStreamWriteLogsTestSuite) getLogsFromCH(logStreamID string) []app.OtelLogRecord {
	var logs []app.OtelLogRecord
	err := s.service.CHDB.
		Where("log_stream_id = ?", logStreamID).
		Order("timestamp ASC").
		Find(&logs).Error
	require.NoError(s.T(), err)
	return logs
}

func (s *LogStreamWriteLogsTestSuite) TestLogStreamWriteLogs() {
	testCases := []struct {
		name             string
		setupFunc        func() (string, []byte)
		expectedCode     int
		expectedLogCount int
		validateFunc     func(string)
	}{
		{
			name: "successfully writes single log",
			setupFunc: func() (string, []byte) {
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

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(logStream)
				})

				body := s.buildOTLPLogExportRequest([]string{"Single log message"})
				return logStream.ID, body
			},
			expectedCode:     http.StatusCreated,
			expectedLogCount: 1,
			validateFunc: func(logStreamID string) {
				logs := s.getLogsFromCH(logStreamID)
				assert.Len(s.T(), logs, 1)
				assert.Equal(s.T(), "Single log message", logs[0].Body)
				assert.Equal(s.T(), "Info", logs[0].SeverityText)
				assert.Equal(s.T(), "test-service", logs[0].ServiceName)
			},
		},
		{
			name: "successfully writes multiple logs in batch",
			setupFunc: func() (string, []byte) {
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

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(logStream)
				})

				messages := []string{"Log 1", "Log 2", "Log 3", "Log 4", "Log 5"}
				body := s.buildOTLPLogExportRequest(messages)
				return logStream.ID, body
			},
			expectedCode:     http.StatusCreated,
			expectedLogCount: 5,
			validateFunc: func(logStreamID string) {
				logs := s.getLogsFromCH(logStreamID)
				assert.Len(s.T(), logs, 5)

				// Verify log bodies
				expectedMessages := []string{"Log 1", "Log 2", "Log 3", "Log 4", "Log 5"}
				for i, log := range logs {
					assert.Equal(s.T(), expectedMessages[i], log.Body)
				}
			},
		},
		{
			name: "validates log stream exists",
			setupFunc: func() (string, []byte) {
				// Use non-existent log stream ID
				body := s.buildOTLPLogExportRequest([]string{"Test log"})
				return "lgsnonexistent123456789012", body
			},
			expectedCode: http.StatusNotFound,
			validateFunc: func(logStreamID string) {
				// No logs should be written
				count := s.countLogsInCH(logStreamID)
				assert.Equal(s.T(), 0, count)
			},
		},
		{
			name: "writes logs to log stream in different org (runner routes are not org-scoped)",
			setupFunc: func() (string, []byte) {
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

				body := s.buildOTLPLogExportRequest([]string{"Test log"})
				// Runner routes don't validate org ownership
				return logStream2.ID, body
			},
			expectedCode:     http.StatusCreated,
			expectedLogCount: 1,
			validateFunc: func(logStreamID string) {
				count := s.countLogsInCH(logStreamID)
				assert.Equal(s.T(), 1, count)
			},
		},
		{
			name: "rejects invalid protobuf payload",
			setupFunc: func() (string, []byte) {
				body := s.buildInvalidOTLPRequest()
				return s.testLogStream.ID, body
			},
			expectedCode: http.StatusInternalServerError,
			validateFunc: func(logStreamID string) {
				// No logs should be written
				count := s.countLogsInCH(logStreamID)
				assert.Equal(s.T(), 0, count)
			},
		},
		{
			name: "writes logs with resource attributes",
			setupFunc: func() (string, []byte) {
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

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(logStream)
				})

				body := s.buildOTLPLogExportRequest([]string{"Log with attributes"})
				return logStream.ID, body
			},
			expectedCode:     http.StatusCreated,
			expectedLogCount: 1,
			validateFunc: func(logStreamID string) {
				logs := s.getLogsFromCH(logStreamID)
				require.Len(s.T(), logs, 1)

				// Verify resource attributes
				assert.NotNil(s.T(), logs[0].ResourceAttributes)
				assert.Equal(s.T(), "test-service", logs[0].ResourceAttributes["service.name"])
				assert.Equal(s.T(), "rgrp123456789", logs[0].ResourceAttributes["runner_group.id"])

				// Verify log attributes
				assert.NotNil(s.T(), logs[0].LogAttributes)
				assert.Equal(s.T(), "rnr123456789", logs[0].LogAttributes["runner.id"])
				assert.Equal(s.T(), "rjb123456789", logs[0].LogAttributes["runner_job.id"])
			},
		},
		{
			name: "writes logs with parent log stream",
			setupFunc: func() (string, []byte) {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create parent log stream
				parentLogStream := &app.LogStream{
					ID:      domains.NewLogStreamID(),
					OrgID:   s.testOrg.ID,
					OwnerID: s.testOrg.ID,
					Open:    true,
				}
				err := s.service.DB.WithContext(ctx).Create(parentLogStream).Error
				require.NoError(s.T(), err)

				// Create child log stream with parent
				childLogStream := &app.LogStream{
					ID:                domains.NewLogStreamID(),
					OrgID:             s.testOrg.ID,
					OwnerID:           s.testOrg.ID,
					Open:              true,
					ParentLogStreamID: generics.NewNullString(parentLogStream.ID),
				}
				err = s.service.DB.WithContext(ctx).Create(childLogStream).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(childLogStream)
					s.service.DB.Unscoped().Delete(parentLogStream)
				})

				body := s.buildOTLPLogExportRequest([]string{"Log with parent"})
				return childLogStream.ID, body
			},
			expectedCode:     http.StatusCreated,
			expectedLogCount: 1, // countLogsInCH counts by child log stream ID only
			validateFunc: func(logStreamID string) {
				// Get the child log stream to find parent
				var childLogStream app.LogStream
				err := s.service.DB.Where("id = ?", logStreamID).First(&childLogStream).Error
				require.NoError(s.T(), err)

				// Verify logs written to child log stream
				childLogs := s.getLogsFromCH(logStreamID)
				assert.Len(s.T(), childLogs, 1)
				assert.Equal(s.T(), "Log with parent", childLogs[0].Body)

				// Verify logs also written to parent log stream
				parentLogs := s.getLogsFromCH(childLogStream.ParentLogStreamID.String)
				assert.Len(s.T(), parentLogs, 1)
				assert.Equal(s.T(), "Log with parent", parentLogs[0].Body)
			},
		},
		{
			name: "empty request body returns error",
			setupFunc: func() (string, []byte) {
				return s.testLogStream.ID, []byte{}
			},
			expectedCode: http.StatusInternalServerError,
			validateFunc: func(logStreamID string) {
				// No logs should be written
				count := s.countLogsInCH(logStreamID)
				assert.Equal(s.T(), 0, count)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			logStreamID, body := tc.setupFunc()
			path := fmt.Sprintf("/v1/log-streams/%s/logs", logStreamID)
			rr := s.makeRequest("POST", path, body)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusCreated {
				// Verify response
				var response string
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(s.T(), err)
				assert.Equal(s.T(), "ok", response)

				// Verify logs were written to ClickHouse
				if tc.expectedLogCount > 0 {
					// Give ClickHouse a moment to process the write
					time.Sleep(100 * time.Millisecond)

					count := s.countLogsInCH(logStreamID)
					assert.Equal(s.T(), tc.expectedLogCount, count,
						"Expected %d logs in ClickHouse, got %d", tc.expectedLogCount, count)
				}
			}

			if tc.validateFunc != nil {
				// Give ClickHouse a moment to process
				time.Sleep(100 * time.Millisecond)
				tc.validateFunc(logStreamID)
			}
		})
	}
}
