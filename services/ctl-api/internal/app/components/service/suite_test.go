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
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	tclient "go.temporal.io/sdk/client"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	temporal "github.com/nuonco/nuon/pkg/temporal/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type mockWorkflowRun struct{}

func (m *mockWorkflowRun) GetID() string    { return "mock-workflow-id" }
func (m *mockWorkflowRun) GetRunID() string { return "mock-run-id" }
func (m *mockWorkflowRun) Get(ctx context.Context, valuePtr interface{}) error {
	return nil
}
func (m *mockWorkflowRun) GetWithOptions(ctx context.Context, valuePtr interface{}, options tclient.WorkflowRunGetOptions) error {
	return nil
}

type ComponentsTestDeps struct {
	fx.In

	DB     *gorm.DB `name:"psql"`
	CHDB   *gorm.DB `name:"ch"`
	V      *validator.Validate
	L      *zap.Logger
	MW     metrics.Writer
	Seeder *testseed.Seeder
}

type ComponentsServiceTestSuite struct {
	tests.BaseDBTestSuite

	fxApp             *fxtest.App
	deps              ComponentsTestDeps
	componentsService *service
	router            *gin.Engine
	ctx               context.Context
	testOrg           *app.Org
	testAcc           *app.Account
	testApp           *app.App
	testAppConfig     *app.AppConfig
	mockTC            *temporal.MockClient
}

func TestComponentsServiceSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(ComponentsServiceTestSuite))
}

func (s *ComponentsServiceTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(s.T())
	s.mockTC = temporal.NewMockClient(ctrl)

	s.mockTC.EXPECT().ExecuteWorkflowInNamespace(
		gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
	).Return(&mockWorkflowRun{}, nil).AnyTimes()

	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T: s.T(),
			Mocks: &tests.TestMocks{
				MockTC: s.mockTC,
			},
			CustomValidator: true,
		}),
		fx.Provide(New),
		fx.Populate(&s.deps, &s.componentsService),
	)

	s.fxApp = fxtest.New(s.T(), options...)
	s.fxApp.RequireStart()

	s.SetDB(s.deps.DB)
}

func (s *ComponentsServiceTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()

	s.setupTestData()

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.deps.L,
		DB:      s.deps.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := s.componentsService.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *ComponentsServiceTestSuite) TearDownSuite() {
	s.fxApp.RequireStop()
}

func (s *ComponentsServiceTestSuite) setupTestData() {
	s.ctx = context.Background()
	s.ctx, s.testAcc = s.deps.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.deps.Seeder.EnsureOrg(s.ctx, s.T())
	s.testApp = s.deps.Seeder.CreateApp(s.ctx, s.T())
	s.testAppConfig = s.deps.Seeder.CreateAppConfig(s.ctx, s.T(), s.testApp.ID)
}

func (s *ComponentsServiceTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *ComponentsServiceTestSuite) makeRawRequest(method, path string, rawBody string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, bytes.NewBufferString(rawBody))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

func (s *ComponentsServiceTestSuite) getSeededComponent(componentType app.ComponentType) *app.Component {
	for _, ccc := range s.testAppConfig.ComponentConfigConnections {
		var cmp app.Component
		res := s.deps.DB.First(&cmp, "id = ?", ccc.ComponentID)
		if res.Error == nil && cmp.Type == componentType {
			return &cmp
		}
	}
	s.T().Fatalf("no seeded component of type %s", componentType)
	return nil
}

func (s *ComponentsServiceTestSuite) getSeededConfigConnection(componentID string) *app.ComponentConfigConnection {
	for _, ccc := range s.testAppConfig.ComponentConfigConnections {
		if ccc.ComponentID == componentID {
			return &ccc
		}
	}
	s.T().Fatalf("no seeded config connection for component %s", componentID)
	return nil
}
