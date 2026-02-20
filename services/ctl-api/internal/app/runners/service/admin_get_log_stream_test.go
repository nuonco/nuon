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

type AdminGetLogStreamTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type AdminGetLogStreamTestSuite struct {
	tests.BaseDBTestSuite
	app           *fxtest.App
	service       AdminGetLogStreamTestService
	router        *gin.Engine
	testOrg       *app.Org
	testAcc       *app.Account
	testLogStream *app.LogStream
}

func TestAdminGetLogStreamSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(AdminGetLogStreamTestSuite))
}

func (s *AdminGetLogStreamTestSuite) SetupSuite() {
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

func (s *AdminGetLogStreamTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// Create router with internal routes (no org context for admin routes)
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:  s.service.L,
		DB: s.service.DB,
	})
	err := s.service.RunnersService.RegisterInternalRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *AdminGetLogStreamTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AdminGetLogStreamTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	s.testOrg = s.service.Seeder.CreateOrg(ctx, s.T())

	// Create log stream
	s.testLogStream = &app.LogStream{
		ID:        domains.NewLogStreamID(),
		OrgID:     s.testOrg.ID,
		OwnerID:   s.testOrg.ID,
		OwnerType: "org",
		Open:      true,
	}
	err := s.service.DB.WithContext(ctx).Create(s.testLogStream).Error
	require.NoError(s.T(), err)
}

func (s *AdminGetLogStreamTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *AdminGetLogStreamTestSuite) TestAdminGetLogStream() {
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
				return s.testLogStream.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(ls *app.LogStream) {
				assert.Equal(s.T(), s.testLogStream.ID, ls.ID)
				assert.Equal(s.T(), s.testOrg.ID, ls.OrgID)
				assert.Equal(s.T(), s.testOrg.ID, ls.OwnerID)
				assert.Equal(s.T(), "org", ls.OwnerType)
				assert.True(s.T(), ls.Open)
			},
		},
		{
			name: "successfully get log stream by owner_id",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				ownerID := domains.NewBuildID()
				ls := &app.LogStream{
					ID:        domains.NewLogStreamID(),
					OrgID:     s.testOrg.ID,
					OwnerID:   ownerID,
					OwnerType: "component_build",
					Open:      true,
				}
				err := s.service.DB.WithContext(ctx).Create(ls).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(ls)
				})

				return ownerID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(ls *app.LogStream) {
				assert.Equal(s.T(), "component_build", ls.OwnerType)
				assert.True(s.T(), ls.Open)
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
			name: "log stream with closed state",
			setupFunc: func() string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)

				ls := &app.LogStream{
					ID:        domains.NewLogStreamID(),
					OrgID:     s.testOrg.ID,
					OwnerID:   s.testOrg.ID,
					OwnerType: "org",
					Open:      false,
				}
				err := s.service.DB.WithContext(ctx).Create(ls).Error
				require.NoError(s.T(), err)

				s.T().Cleanup(func() {
					s.service.DB.Unscoped().Delete(ls)
				})

				return ls.ID
			},
			expectedCode: http.StatusOK,
			validateFunc: func(ls *app.LogStream) {
				assert.False(s.T(), ls.Open)
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
				var ls app.LogStream
				err := json.Unmarshal(rr.Body.Bytes(), &ls)
				require.NoError(s.T(), err)
				tc.validateFunc(&ls)
			}
		})
	}
}
