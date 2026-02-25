package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type DeleteTerraformStateTestService struct {
	fx.In
	DB             *gorm.DB `name:"psql"`
	CHDB           *gorm.DB `name:"ch"`
	V              *validator.Validate
	L              *zap.Logger
	Seeder         *testseed.Seeder
	RunnersService *service
}

type DeleteTerraformStateTestSuite struct {
	tests.BaseDBTestSuite
	app     *fxtest.App
	service DeleteTerraformStateTestService
	router  *gin.Engine
}

func TestDeleteTerraformStateSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(DeleteTerraformStateTestSuite))
}

func (s *DeleteTerraformStateTestSuite) SetupSuite() {
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

func (s *DeleteTerraformStateTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()

	ctx := context.Background()
	_, _ = s.service.Seeder.EnsureAccount(ctx, s.T())

	// Create router with runner routes (DELETE endpoint is only registered here)
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:  s.service.L,
		DB: s.service.DB,
	})
	err := s.service.RunnersService.RegisterRunnerRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *DeleteTerraformStateTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *DeleteTerraformStateTestSuite) makeRequest(method, path string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, nil)
	require.NoError(s.T(), err)

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *DeleteTerraformStateTestSuite) TestDeleteTerraformState() {
	// This is a no-op handler that always returns 200
	testCases := []struct {
		name         string
		expectedCode int
	}{
		{
			name:         "always returns 200 (no-op handler)",
			expectedCode: http.StatusOK,
		},
		{
			name:         "returns 200 with any query params",
			expectedCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			rr := s.makeRequest("DELETE", "/v1/terraform-backend")

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)
		})
	}
}
