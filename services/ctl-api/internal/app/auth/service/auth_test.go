package service

import (
	"context"
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

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type AuthTestService struct {
	fx.In
	DB          *gorm.DB `name:"psql"`
	CHDB        *gorm.DB `name:"ch"`
	V           *validator.Validate
	L           *zap.Logger
	Cfg         *internal.Config
	AuthService *service
	Seeder      *testseed.Seeder
}

type AuthTestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service AuthTestService
	router  *gin.Engine
	testAcc *app.Account
	testOrg *app.Org
}

func TestAuthSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(AuthTestSuite))
}

func (s *AuthTestSuite) SetupSuite() {
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
}

func (s *AuthTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := s.service.AuthService.RegisterAuthRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *AuthTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AuthTestSuite) setupTestData() {
	ctx := context.Background()
	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	_, s.testOrg = s.service.Seeder.EnsureOrg(ctx, s.T())
}

func (s *AuthTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *AuthTestSuite) TestAuth() {
	testCases := []struct {
		name           string
		queryParams    string
		expectedCode   int
		validateFunc   func(*httptest.ResponseRecorder)
		expectedError  bool
		errorSubstring string
	}{
		{
			name:           "missing state parameter",
			queryParams:    "",
			expectedCode:   http.StatusBadRequest,
			expectedError:  true,
			errorSubstring: "missing state parameter",
		},
		{
			name:           "OAuth provider error",
			queryParams:    "?error=access_denied&error_description=User%20denied%20access",
			expectedCode:   http.StatusUnauthorized,
			expectedError:  true,
			errorSubstring: "access_denied",
		},
		{
			name:         "valid state redirects to auth/:state",
			queryParams:  "?state=test-state-12345&code=test-auth-code",
			expectedCode: http.StatusFound,
			validateFunc: func(rr *httptest.ResponseRecorder) {
				location := rr.Header().Get("Location")
				assert.NotEmpty(s.T(), location, "should have Location header")
				assert.Contains(s.T(), location, "/auth/test-state-12345",
					"should redirect to /auth/:state endpoint")
				assert.Contains(s.T(), location, "state=test-state-12345",
					"should preserve state in query")
				assert.Contains(s.T(), location, "code=test-auth-code",
					"should preserve code in query")
			},
		},
		{
			name:         "preserves full query string in redirect",
			queryParams:  "?state=abc123&code=xyz789&extra=param",
			expectedCode: http.StatusFound,
			validateFunc: func(rr *httptest.ResponseRecorder) {
				location := rr.Header().Get("Location")
				assert.Contains(s.T(), location, "/auth/abc123")
				assert.Contains(s.T(), location, "state=abc123")
				assert.Contains(s.T(), location, "code=xyz789")
				assert.Contains(s.T(), location, "extra=param")
			},
		},
		{
			name:           "error with description",
			queryParams:    "?error=server_error&error_description=Internal%20server%20error",
			expectedCode:   http.StatusUnauthorized,
			expectedError:  true,
			errorSubstring: "server_error",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			rr := s.makeRequest("GET", "/auth"+tc.queryParams)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedError {
				body := rr.Body.String()
				assert.Contains(s.T(), body, tc.errorSubstring)
			}

			if tc.validateFunc != nil {
				tc.validateFunc(rr)
			}
		})
	}
}

func (s *AuthTestSuite) TestAuthWithVariousErrors() {
	errorCases := []struct {
		error       string
		description string
	}{
		{"access_denied", "User denied access"},
		{"invalid_request", "Missing required parameter"},
		{"unauthorized_client", "Client not authorized"},
		{"unsupported_response_type", "Response type not supported"},
		{"invalid_scope", "Invalid scope requested"},
		{"server_error", "Internal server error"},
		{"temporarily_unavailable", "Service temporarily unavailable"},
	}

	for _, ec := range errorCases {
		s.Run("error_"+ec.error, func() {
			queryParams := "?error=" + ec.error + "&error_description=" + ec.description
			rr := s.makeRequest("GET", "/auth"+queryParams)

			require.Equal(s.T(), http.StatusUnauthorized, rr.Code)

			body := rr.Body.String()
			assert.Contains(s.T(), body, ec.error)
		})
	}
}
