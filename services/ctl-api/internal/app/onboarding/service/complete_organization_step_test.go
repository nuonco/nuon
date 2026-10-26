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

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	tclient "go.temporal.io/sdk/client"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	temporal "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	createorg "github.com/nuonco/nuon/services/ctl-api/internal/app/onboarding/signals/create_org"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type onboardingTestDeps struct {
	fx.In

	DB                *gorm.DB `name:"psql"`
	L                 *zap.Logger
	OnboardingService *service
	Seeder            *testseed.Seeder
}

type CompleteOrganizationStepTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	service onboardingTestDeps
}

func TestCompleteOrganizationStepSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(CompleteOrganizationStepTestSuite))
}

type mockWorkflowRun struct{}

func (m *mockWorkflowRun) GetID() string                                { return "mock-workflow-id" }
func (m *mockWorkflowRun) GetRunID() string                             { return "mock-run-id" }
func (m *mockWorkflowRun) Get(ctx context.Context, v interface{}) error { return nil }
func (m *mockWorkflowRun) GetWithOptions(ctx context.Context, v interface{}, o tclient.WorkflowRunGetOptions) error {
	return nil
}

func (s *CompleteOrganizationStepTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	mockTC := temporal.NewMockClient(gomock.NewController(s.T()))
	mockTC.EXPECT().ExecuteWorkflowInNamespace(
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
	).Return(&mockWorkflowRun{}, nil).AnyTimes()

	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T:               s.T(),
			Mocks:           &tests.TestMocks{MockTC: mockTC},
			CustomValidator: true,
		}),
		fx.Provide(NewCatalog),
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()

	s.SetDB(s.service.DB)
}

func (s *CompleteOrganizationStepTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	tests.ClearQueueSignals(s.T(), s.service.DB)
}

func (s *CompleteOrganizationStepTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *CompleteOrganizationStepTestSuite) post(router *gin.Engine, path string, body any) *httptest.ResponseRecorder {
	jsonBytes, err := json.Marshal(body)
	require.NoError(s.T(), err)
	req, err := http.NewRequest(http.MethodPost, path, bytes.NewBuffer(jsonBytes))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}

func (s *CompleteOrganizationStepTestSuite) TestCreateNewOrgRejectsFederatedToken() {
	testCases := map[string]struct {
		tokenType      app.TokenType
		wantStatus     int
		wantSignal     bool
		wantStepStatus app.OnboardingStepStatus
	}{
		"federated token": {
			tokenType:  app.TokenTypeFederated,
			wantStatus: http.StatusForbidden,
		},
		"user token": {
			tokenType:      app.TokenTypeNuon,
			wantStatus:     http.StatusOK,
			wantSignal:     true,
			wantStepStatus: app.OnboardingStepStatusInProgress,
		},
	}

	for name, tc := range testCases {
		s.Run(name, func() {
			_, acct := s.service.Seeder.EnsureAccount(context.Background(), s.T())
			router := tests.NewTestRouter(tests.RouterOptions{
				L:       s.service.L,
				DB:      s.service.DB,
				TestAcc: acct,
				AdditionalMiddlewares: []gin.HandlerFunc{
					func(c *gin.Context) {
						cctx.SetTokenTypeGinContext(c, tc.tokenType)
						c.Next()
					},
				},
			})
			require.NoError(s.T(), s.service.OnboardingService.RegisterPublicRoutes(router))

			rr := s.post(router, "/v1/onboarding", nil)
			require.Equal(s.T(), http.StatusCreated, rr.Code, rr.Body.String())
			var onboarding app.Onboarding
			require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &onboarding))

			rr = s.post(router, "/v1/onboarding/current/steps/organization", CompleteOrganizationStepRequest{
				Name: fmt.Sprintf("acme-%s", domains.NewOrgID()[:8]),
			})
			require.Equal(s.T(), tc.wantStatus, rr.Code, rr.Body.String())
			if tc.wantStatus == http.StatusForbidden {
				require.Contains(s.T(), rr.Body.String(), "federated tokens cannot create organizations")
			}

			foundSignal := false
			for _, qs := range tests.GetQueueSignalsByOwner(s.T(), s.service.DB, onboarding.ID) {
				if qs.Type == createorg.SignalType {
					foundSignal = true
				}
			}
			require.Equal(s.T(), tc.wantSignal, foundSignal)

			var stored app.Onboarding
			require.NoError(s.T(), s.service.DB.Where(app.Onboarding{ID: onboarding.ID}).First(&stored).Error)
			require.Equal(s.T(), tc.wantStepStatus, stored.StepStatus)
		})
	}
}
