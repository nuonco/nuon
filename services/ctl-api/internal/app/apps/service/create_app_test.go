package service

import (
	"bytes"
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

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	accountshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/helpers"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type CreateAppTestService struct {
	fx.In

	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	MW              metrics.Writer
	VcsHelpers      *vcshelpers.Helpers
	AppsHelpers     *appshelpers.Helpers
	InstallsHelpers *installshelpers.Helpers
	AccountsHelpers *accountshelpers.Helpers
	AppsService     *service
	Seeder          *testseed.Seeder
}

type CreateAppTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	service CreateAppTestService
	router  *gin.Engine
	ctx     context.Context
	testOrg *app.Org
	testAcc *app.Account
}

func TestCreateAppSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(CreateAppTestSuite))
}

func (s *CreateAppTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T: s.T(),

			CustomValidator: true,
		}),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()

	s.SetDB(s.service.DB)
}

func (s *CreateAppTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := s.service.AppsService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *CreateAppTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *CreateAppTestSuite) setupTestData() {
	s.ctx = context.Background()
	s.ctx, s.testAcc = s.service.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.service.Seeder.EnsureOrg(s.ctx, s.T())
}

func (s *CreateAppTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		require.NoError(s.T(), err)
		reqBody = bytes.NewBuffer(jsonBytes)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, path, reqBody)
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *CreateAppTestSuite) TestCreateAppSuccess() {
	req := CreateAppRequest{
		Name:        "test-app",
		Description: "Test app",
	}
	rr := s.makeRequest(http.MethodPost, "/v1/apps", req)

	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	var response models.AppApp
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	assert.NotEmpty(s.T(), response.ID)
	assert.Equal(s.T(), "test-app", response.Name)
	assert.Equal(s.T(), s.testOrg.ID, response.OrgID)

	var dbApp app.App
	err = s.service.DB.First(&dbApp, "id = ?", response.ID).Error
	require.NoError(s.T(), err)
	assert.Equal(s.T(), "test-app", dbApp.Name)
	assert.Equal(s.T(), s.testOrg.ID, dbApp.OrgID)
}

func (s *CreateAppTestSuite) TestCreateAppValidationError() {
	testCases := []struct {
		appName  string
		testName string
	}{
		{appName: "", testName: "empty name"},
		{appName: "my app", testName: "name with spaces"},
		{appName: "MyApp", testName: "name with uppercase"},
		{appName: "my-app!@#", testName: "name with special chars"},
		{appName: "my.app", testName: "name with dots"},
		{appName: "my/app", testName: "name with slashes"},
	}

	for _, tc := range testCases {
		s.Run(tc.testName, func() {
			req := CreateAppRequest{
				Name: tc.appName,
			}
			rr := s.makeRequest(http.MethodPost, "/v1/apps", req)

			if rr.Code != http.StatusBadRequest {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusBadRequest, rr.Code)
		})
	}
}

func (s *CreateAppTestSuite) TestCreateAppDuplicateName() {
	s.Run("within org", func() {
		existingApp := s.service.Seeder.CreateApp(s.ctx, s.T())

		req := CreateAppRequest{Name: existingApp.Name}
		rr := s.makeRequest(http.MethodPost, "/v1/apps", req)

		if rr.Code != http.StatusConflict {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusConflict, rr.Code)
	})

	s.Run("across orgs", func() {
		ctx2 := context.Background()
		ctx2, _ = s.service.Seeder.EnsureAccount(ctx2, s.T())
		ctx2, _ = s.service.Seeder.EnsureOrg(ctx2, s.T())
		existingApp := s.service.Seeder.CreateApp(ctx2, s.T())

		req := CreateAppRequest{Name: existingApp.Name}
		rr := s.makeRequest(http.MethodPost, "/v1/apps", req)

		if rr.Code != http.StatusCreated {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusCreated, rr.Code)
	})
}
