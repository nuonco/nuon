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

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	executeflow "github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type ReprovisionInstallTestService struct {
	fx.In
	DB              *gorm.DB `name:"psql"`
	CHDB            *gorm.DB `name:"ch"`
	V               *validator.Validate
	L               *zap.Logger
	Seeder          *testseed.Seeder
	InstallsService *service
}

type ReprovisionInstallTestSuite struct {
	tests.BaseDBTestSuite
	app         *fxtest.App
	service     ReprovisionInstallTestService
	router      *gin.Engine
	testOrg     *app.Org
	testAcc     *app.Account
	testApp     *app.App
	testInstall *app.Install
}

func TestReprovisionInstallSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(ReprovisionInstallTestSuite))
}

func (s *ReprovisionInstallTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T:               s.T(),
			CustomValidator: true,
		}),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.service.DB)
}

func (s *ReprovisionInstallTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	// ReprovisionInstall creates install_workflows which require created_by_id and org_id.
	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})
	err := s.service.InstallsService.RegisterInternalRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *ReprovisionInstallTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *ReprovisionInstallTestSuite) setupTestData() {
	ctx := context.Background()

	ctx, s.testAcc = s.service.Seeder.EnsureAccount(ctx, s.T())
	ctx, s.testOrg = s.service.Seeder.EnsureOrg(ctx, s.T())
	s.testApp = s.service.Seeder.CreateApp(ctx, s.T())
	s.service.Seeder.CreateAppConfig(ctx, s.T(), s.testApp.ID)
	s.testInstall = s.service.Seeder.CreateInstall(ctx, s.T(), s.testApp)
}

func (s *ReprovisionInstallTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		require.NoError(s.T(), err)
	}

	req, err := http.NewRequest(method, path, bytes.NewBuffer(reqBody))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *ReprovisionInstallTestSuite) TestReprovisionInstall() {
	testCases := []struct {
		name             string
		setupFunc        func() string
		requestBody      interface{}
		expectedCode     int
		expectedSignal   bool
		validateFunc     func(string)
		expectedNotFound bool
	}{
		{
			name: "successfully reprovision install",
			setupFunc: func() string {
				return s.testInstall.ID
			},
			requestBody: ReprovisionInstallRequest{
				PlanOnly: false,
			},
			expectedCode:   http.StatusCreated,
			expectedSignal: true,
			validateFunc: func(installID string) {
				// Verify workflow was created
				var workflow app.Workflow
				err := s.service.DB.Where("owner_id = ?", installID).
					Where("owner_type = ?", "installs").
					Where("type = ?", app.WorkflowTypeReprovision).
					First(&workflow).Error
				require.NoError(s.T(), err)
				assert.Equal(s.T(), app.WorkflowTypeReprovision, workflow.Type)
				assert.False(s.T(), workflow.PlanOnly)

				// Verify signal contains workflow ID
				sigs := tests.GetQueueSignals(s.T(), s.service.DB)
				require.Len(s.T(), sigs, 1)
				assert.Equal(s.T(), executeflow.SignalType, sigs[0].Type)
			},
		},
		{
			name: "reprovision with plan only mode",
			setupFunc: func() string {
				return s.testInstall.ID
			},
			requestBody: ReprovisionInstallRequest{
				PlanOnly: true,
			},
			expectedCode:   http.StatusCreated,
			expectedSignal: true,
			validateFunc: func(installID string) {
				// Verify workflow was created with plan only
				var workflow app.Workflow
				err := s.service.DB.Where("owner_id = ?", installID).
					Where("owner_type = ?", "installs").
					Where("type = ?", app.WorkflowTypeReprovision).
					Order("created_at DESC").
					First(&workflow).Error
				require.NoError(s.T(), err)
				assert.True(s.T(), workflow.PlanOnly, "workflow should be plan only")
			},
		},
		{
			name: "reprovision with empty body defaults to no plan only",
			setupFunc: func() string {
				return s.testInstall.ID
			},
			requestBody:    nil,
			expectedCode:   http.StatusCreated,
			expectedSignal: true,
			validateFunc: func(installID string) {
				var workflow app.Workflow
				err := s.service.DB.Where("owner_id = ?", installID).
					Where("owner_type = ?", "installs").
					Where("type = ?", app.WorkflowTypeReprovision).
					Order("created_at DESC").
					First(&workflow).Error
				require.NoError(s.T(), err)
				assert.False(s.T(), workflow.PlanOnly)
			},
		},
		{
			name: "nonexistent install returns error",
			setupFunc: func() string {
				return "ins000000000000000000000000"
			},
			requestBody: ReprovisionInstallRequest{
				PlanOnly: false,
			},
			expectedCode:     http.StatusNotFound,
			expectedSignal:   false,
			expectedNotFound: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			installID := tc.setupFunc()
			rr := s.makeRequest("POST", "/v1/installs/"+installID+"/admin-reprovision", tc.requestBody)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedNotFound {
				assert.Contains(s.T(), rr.Body.String(), "error")
			}

			if tc.validateFunc != nil && rr.Code == http.StatusCreated {
				tc.validateFunc(installID)
			}

			// Verify signal presence matches expectation
			allSignals := tests.GetQueueSignals(s.T(), s.service.DB)
			if tc.expectedSignal {
				assert.GreaterOrEqual(s.T(), len(allSignals), 1, "expected signal to be sent")
			} else {
				assert.Len(s.T(), allSignals, 0, "expected no signal to be sent")
			}
		})
	}
}
