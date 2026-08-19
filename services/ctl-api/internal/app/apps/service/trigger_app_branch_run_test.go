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
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	accountshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/helpers"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

// TriggerAppBranchRunTestService holds all fx-injected dependencies for
// TriggerAppBranchRun endpoint tests.
type TriggerAppBranchRunTestService struct {
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

// TriggerAppBranchRunTestSuite is the testify suite for the TriggerAppBranchRun endpoint.
//
// Only the request-validation rejections are covered here: a full happy-path
// run enqueues Temporal signals via the queue/workflow machinery, which is out
// of scope for this suite.
type TriggerAppBranchRunTestSuite struct {
	tests.BaseDBTestSuite

	fxApp      *fxtest.App
	service    TriggerAppBranchRunTestService
	router     *gin.Engine
	ctx        context.Context
	testOrg    *app.Org
	testAcc    *app.Account
	testApp    *app.App
	testBranch *app.AppBranch
}

func TestTriggerAppBranchRunSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(TriggerAppBranchRunTestSuite))
}

func (s *TriggerAppBranchRunTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(s.T()),
		// service under test
		fx.Provide(New),
		fx.Populate(&s.service),
	)

	s.fxApp = fxtest.New(s.T(), options...)
	s.fxApp.RequireStart()

	// Store DB reference for automatic truncation
	s.SetDB(s.service.DB)
}

func (s *TriggerAppBranchRunTestSuite) SetupTest() {
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

func (s *TriggerAppBranchRunTestSuite) TearDownSuite() {
	s.fxApp.RequireStop()
}

func (s *TriggerAppBranchRunTestSuite) setupTestData() {
	s.ctx = context.Background()
	s.ctx, s.testAcc = s.service.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.service.Seeder.EnsureOrg(s.ctx, s.T())
	s.testApp = s.service.Seeder.CreateApp(s.ctx, s.T())

	s.testBranch = &app.AppBranch{
		ID:          domains.NewAppBranchID(),
		OrgID:       s.testOrg.ID,
		AppID:       s.testApp.ID,
		CreatedByID: s.testAcc.ID,
		Name:        "test-branch",
		ManagedBy:   app.AppBranchManagedByManually,
	}
	require.NoError(s.T(), s.service.DB.WithContext(s.ctx).Create(s.testBranch).Error)
}

func (s *TriggerAppBranchRunTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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

func (s *TriggerAppBranchRunTestSuite) makeRawRequest(method, path, rawBody string) *httptest.ResponseRecorder {
	req, err := http.NewRequest(method, path, bytes.NewBufferString(rawBody))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

// TestTriggerAppBranchRunRejectsInvalidSyncAppConfig covers the two new
// cross-field rejections on TriggerAppBranchRunRequest.Validate. Neither case
// reaches the branch/config lookups or the helper that enqueues the run, so
// no Temporal/queue mocking is required.
func (s *TriggerAppBranchRunTestSuite) TestTriggerAppBranchRunRejectsInvalidSyncAppConfig() {
	testCases := []struct {
		name          string
		req           TriggerAppBranchRunRequest
		expectContain string
	}{
		{
			name: "sync_app_config without app_config_id",
			req: TriggerAppBranchRunRequest{
				SyncAppConfig: true,
			},
			expectContain: "sync_app_config requires app_config_id",
		},
		{
			name: "sync_app_config combined with skip_builds",
			req: TriggerAppBranchRunRequest{
				SyncAppConfig: true,
				AppConfigID:   "cfgplaceholder0000000000000",
				SkipBuilds:    true,
			},
			expectContain: "sync_app_config cannot be combined with skip_builds",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			path := fmt.Sprintf("/v1/apps/%s/branches/%s/runs", s.testApp.ID, s.testBranch.ID)
			rr := s.makeRequest(http.MethodPost, path, tc.req)

			if rr.Code != http.StatusBadRequest {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusBadRequest, rr.Code)
			assert.Contains(s.T(), rr.Body.String(), tc.expectContain)

			var count int64
			require.NoError(s.T(), s.service.DB.WithContext(s.ctx).
				Model(&app.AppBranchRun{}).
				Where(app.AppBranchRun{AppBranchID: s.testBranch.ID}).
				Count(&count).Error)
			assert.EqualValues(s.T(), 0, count, "expected no run to be created for an invalid request")
		})
	}
}
