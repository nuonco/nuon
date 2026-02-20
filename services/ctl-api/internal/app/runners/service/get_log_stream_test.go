package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

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

type GetLogStreamTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type GetLogStreamTestSuite struct {
	tests.BaseDBTestSuite
	app           *fxtest.App
	service       GetLogStreamTestService
	router        *gin.Engine
	testOrg       *app.Org
	testAcc       *app.Account
	testLogStream *app.LogStream
}

func TestGetLogStreamSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(GetLogStreamTestSuite))
}

func (s *GetLogStreamTestSuite) SetupSuite() {
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
}

func (s *GetLogStreamTestSuite) SetupTest() {
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

func (s *GetLogStreamTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetLogStreamTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	s.testOrg = s.service.Seeder.CreateOrg(ctx, s.T())

	// Create log stream
	s.testLogStream = &app.LogStream{
		ID:      domains.NewLogStreamID(),
		OrgID:   s.testOrg.ID,
		OwnerID: s.testOrg.ID,
		Open:    true,
	}
	err := s.service.DB.WithContext(ctx).Create(s.testLogStream).Error
	require.NoError(s.T(), err)
}

func (s *GetLogStreamTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *GetLogStreamTestSuite) TestGetLogStream() {
	testCases := []struct {
		name             string
		setupFunc        func() string
		expectedCode     int
		validateFunc     func(*app.LogStream)
		expectedNotFound bool
	}{
		{
			name: "successfully get log stream by ID",
			setupFunc: func() string {
				// Use the existing test log stream
				return s.testLogStream.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(logStream *app.LogStream) {
				assert.Equal(s.T(), s.testLogStream.ID, logStream.ID)
				assert.Equal(s.T(), s.testOrg.ID, logStream.OrgID)
				assert.Equal(s.T(), s.testOrg.ID, logStream.OwnerID)
				assert.True(s.T(), logStream.Open)
			},
		},
		{
			name: "log stream not found returns error",
			setupFunc: func() string {
				return "lgsnonexistent123456789012"
			},
			expectedCode:     http.StatusNotFound,
			expectedNotFound: true,
		},
		{
			name: "log stream in different org not accessible",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create second org
				org2 := &app.Org{
					ID:          domains.NewOrgID(),
					Name:        "other-org",
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
			expectedCode:     http.StatusNotFound,
			expectedNotFound: true,
		},
		{
			name: "multiple log streams exist, correct one returned by ID",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				// Create additional log streams in same org
				logStream2 := &app.LogStream{
					ID:      domains.NewLogStreamID(),
					OrgID:   s.testOrg.ID,
					OwnerID: s.testOrg.ID,
					Open:    false,
				}
				err := s.service.DB.WithContext(ctx).Create(logStream2).Error
				require.NoError(s.T(), err)

				logStream3 := &app.LogStream{
					ID:      domains.NewLogStreamID(),
					OrgID:   s.testOrg.ID,
					OwnerID: s.testOrg.ID,
					Open:    true,
				}
				err = s.service.DB.WithContext(ctx).Create(logStream3).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(logStream2)
					s.service.DB.Unscoped().Delete(logStream3)
				})

				// Return logStream2's ID for verification
				return logStream2.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(logStream *app.LogStream) {
				assert.Equal(s.T(), s.testOrg.ID, logStream.OrgID)
				assert.False(s.T(), logStream.Open, "should be closed")
			},
		},
		{
			name: "verify closed log stream state is preserved",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				logStream := &app.LogStream{
					ID:      domains.NewLogStreamID(),
					OrgID:   s.testOrg.ID,
					OwnerID: s.testOrg.ID,
					Open:    false,
				}
				err := s.service.DB.WithContext(ctx).Create(logStream).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(logStream)
				})

				return logStream.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(logStream *app.LogStream) {
				assert.False(s.T(), logStream.Open)
				assert.Equal(s.T(), s.testOrg.ID, logStream.OrgID)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			logStreamID := tc.setupFunc()
			rr := s.makeRequest("GET", "/v1/log-streams/"+logStreamID)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedNotFound {
				assert.Contains(s.T(), rr.Body.String(), "error")
			} else if tc.validateFunc != nil {
				var logStream app.LogStream
				err := json.Unmarshal(rr.Body.Bytes(), &logStream)
				require.NoError(s.T(), err)
				tc.validateFunc(&logStream)
			}
		})
	}
}

func (s *GetLogStreamTestSuite) TestGetLogStreamOpenClosedStates() {
	states := []struct {
		open        bool
		description string
	}{
		{true, "open"},
		{false, "closed"},
	}

	for _, tc := range states {
		s.Run(tc.description, func() {
			ctx := context.Background()
			ctx = cctx.SetAccountContext(ctx, s.testAcc)

			logStream := &app.LogStream{
				ID:      domains.NewLogStreamID(),
				OrgID:   s.testOrg.ID,
				OwnerID: s.testOrg.ID,
				Open:    tc.open,
			}
			err := s.service.DB.WithContext(ctx).Create(logStream).Error
			require.NoError(s.T(), err)

			rr := s.makeRequest("GET", "/v1/log-streams/"+logStream.ID)

			if rr.Code != http.StatusOK {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusOK, rr.Code)

			var result app.LogStream
			err = json.Unmarshal(rr.Body.Bytes(), &result)
			require.NoError(s.T(), err)
			assert.Equal(s.T(), tc.open, result.Open)

			s.service.DB.Unscoped().Delete(logStream)
		})
	}
}
