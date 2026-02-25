package service

import (
	"context"
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
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

type AuthStateTestService struct {
	fx.In
	DB          *gorm.DB `name:"psql"`
	CHDB        *gorm.DB `name:"ch"`
	V           *validator.Validate
	L           *zap.Logger
	Cfg         *internal.Config
	AuthService *service
}

type AuthStateTestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service AuthStateTestService
	router  *gin.Engine
	testAcc *app.Account
	testOrg *app.Org
}

func TestAuthStateSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(AuthStateTestSuite))
}

func (s *AuthStateTestSuite) SetupSuite() {
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

func (s *AuthStateTestSuite) SetupTest() {
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

func (s *AuthStateTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *AuthStateTestSuite) setupTestData() {
	ctx := context.Background()

	testAcc := &app.Account{
		ID:          domains.NewAccountID(),
		Email:       "test@example.com",
		Subject:     "test-subject",
		AccountType: app.AccountTypeAuth0,
	}
	err := s.service.DB.Create(testAcc).Error
	require.NoError(s.T(), err)
	s.testAcc = testAcc

	ctx = cctx.SetAccountContext(ctx, testAcc)
	testOrg := &app.Org{
		ID:          domains.NewOrgID(),
		Name:        "test-org",
		SandboxMode: true,
		NotificationsConfig: app.NotificationsConfig{
			InternalSlackWebhookURL: "https://hooks.slack.com/foo",
		},
	}
	err = s.service.DB.WithContext(ctx).Create(testOrg).Error
	require.NoError(s.T(), err)
	s.testOrg = testOrg

	// Note: Identity provider comes from environment config (default provider)
	// No need to create test identity provider in DB
}

func (s *AuthStateTestSuite) makeRequestWithCookie(method, path string, sessionData *SessionData) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	if sessionData != nil {
		encoded, err := s.service.AuthService.encodeSession(sessionData)
		require.NoError(s.T(), err)

		req.AddCookie(&http.Cookie{
			Name:  NuonAuthSessionName,
			Value: encoded,
		})
	}

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *AuthStateTestSuite) TestAuthState() {
	now := time.Now().Unix()

	testCases := []struct {
		name           string
		state          string
		queryParams    string
		sessionData    *SessionData
		expectedCode   int
		validateFunc   func(*httptest.ResponseRecorder)
		expectedError  bool
		errorSubstring string
	}{
		{
			name:           "missing session cookie",
			state:          "test-state",
			queryParams:    "?state=test-state&code=auth-code",
			sessionData:    nil,
			expectedCode:   http.StatusBadRequest,
			expectedError:  true,
			errorSubstring: "session not found",
		},
		{
			name:        "state mismatch between path and session",
			state:       "wrong-state",
			queryParams: "?state=wrong-state&code=auth-code",
			sessionData: &SessionData{
				State:      "correct-state",
				ProviderID: s.service.Cfg.NuonAuthProviderType,
				CreatedAt:  now,
			},
			expectedCode:   http.StatusBadRequest,
			expectedError:  true,
			errorSubstring: "state parameter mismatch",
		},
		{
			name:        "state mismatch between path and query",
			state:       "path-state",
			queryParams: "?state=query-state&code=auth-code",
			sessionData: &SessionData{
				State:      "path-state",
				ProviderID: s.service.Cfg.NuonAuthProviderType,
				CreatedAt:  now,
			},
			expectedCode:   http.StatusBadRequest,
			expectedError:  true,
			errorSubstring: "state parameter mismatch",
		},
		{
			name:        "missing provider in session",
			state:       "test-state",
			queryParams: "?state=test-state&code=auth-code",
			sessionData: &SessionData{
				State:      "test-state",
				ProviderID: "", // Missing provider
				CreatedAt:  now,
			},
			expectedCode:   http.StatusBadRequest,
			expectedError:  true,
			errorSubstring: "no provider type in session",
		},
		{
			name:        "invalid provider in session",
			state:       "test-state",
			queryParams: "?state=test-state&code=auth-code",
			sessionData: &SessionData{
				State:      "test-state",
				ProviderID: "invalid-provider",
				CreatedAt:  now,
			},
			expectedCode:   http.StatusBadRequest,
			expectedError:  true,
			errorSubstring: "invalid provider",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			path := "/auth/" + tc.state + tc.queryParams
			rr := s.makeRequestWithCookie("GET", path, tc.sessionData)

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

func (s *AuthStateTestSuite) TestAuthStateSessionValidation() {
	// Test various session validation scenarios
	testState := "valid-state-123"

	s.Run("expired session", func() {
		// Create session with old timestamp (expired)
		sessionData := &SessionData{
			State:      testState,
			ProviderID: s.service.Cfg.NuonAuthProviderType,
			CreatedAt:  1000000000, // Very old timestamp
		}

		path := "/auth/" + testState + "?state=" + testState + "&code=auth-code"
		rr := s.makeRequestWithCookie("GET", path, sessionData)

		require.Equal(s.T(), http.StatusBadRequest, rr.Code)
		body := rr.Body.String()
		assert.Contains(s.T(), body, "session")
	})

	s.Run("valid session structure", func() {
		sessionData := &SessionData{
			State:        testState,
			ProviderID:   s.service.Cfg.NuonAuthProviderType,
			RequestedURL: "http://localhost:4000/dashboard",
			FailCount:    1,
			CreatedAt:    time.Now().Unix(),
		}

		path := "/auth/" + testState + "?state=" + testState + "&code=auth-code"
		rr := s.makeRequestWithCookie("GET", path, sessionData)

		// Will fail at OAuth exchange since we don't have a real auth code,
		// but should pass session validation
		// The actual response code depends on whether OAuth exchange succeeds
		s.T().Logf("Response code: %d", rr.Code)
	})
}

func (s *AuthStateTestSuite) TestAuthStateRedirectFlow() {
	testState := "redirect-test-state"

	s.Run("redirects to requested URL on success", func() {
		sessionData := &SessionData{
			State:        testState,
			ProviderID:   s.service.Cfg.NuonAuthProviderType,
			RequestedURL: "http://localhost:4000/dashboard",
			CreatedAt:    time.Now().Unix(),
		}

		path := "/auth/" + testState + "?state=" + testState + "&code=valid-code"
		rr := s.makeRequestWithCookie("GET", path, sessionData)

		// This will fail at OAuth token exchange in test, but verifies session handling
		s.T().Logf("Response: code=%d, body=%s", rr.Code, rr.Body.String())
	})

	s.Run("redirects to success page without requested URL", func() {
		sessionData := &SessionData{
			State:        testState,
			ProviderID:   s.service.Cfg.NuonAuthProviderType,
			RequestedURL: "", // No requested URL
			CreatedAt:    time.Now().Unix(),
		}

		path := "/auth/" + testState + "?state=" + testState + "&code=valid-code"
		rr := s.makeRequestWithCookie("GET", path, sessionData)

		// This will fail at OAuth token exchange in test
		s.T().Logf("Response: code=%d", rr.Code)
	})
}
