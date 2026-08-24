// Integration tests: run with INTEGRATION=true against the migrated test database.
package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	installshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type stackPhoneHomeDeps struct {
	fx.In

	DB              *gorm.DB `name:"psql"`
	L               *zap.Logger
	V               *validator.Validate
	Seeder          *testseed.Seeder
	InstallsHelpers *installshelpers.Helpers
}

type StackPhoneHomeTestSuite struct {
	tests.BaseDBTestSuite

	fxApp  *fxtest.App
	deps   stackPhoneHomeDeps
	svc    *service
	router *gin.Engine

	ctx      context.Context
	testOrg  *app.Org
	testAcc  *app.Account
	testApp  *app.App
	appCfg   *app.AppConfig
	otherOrg *app.Org
}

func TestStackPhoneHomeSuite(t *testing.T) {
	tests.SkipIfNotIntegration(t)
	suite.Run(t, new(StackPhoneHomeTestSuite))
}

func (s *StackPhoneHomeTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(s.T()),
		fx.Populate(&s.deps),
	)
	s.fxApp = fxtest.New(s.T(), options...)
	s.fxApp.RequireStart()
	s.SetDB(s.deps.DB)
}

func (s *StackPhoneHomeTestSuite) TearDownSuite() {
	s.fxApp.RequireStop()
}

func (s *StackPhoneHomeTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()

	s.ctx = context.Background()
	s.ctx, s.testAcc = s.deps.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.deps.Seeder.EnsureOrg(s.ctx, s.T())
	s.testApp = s.deps.Seeder.CreateApp(s.ctx, s.T())
	s.appCfg = s.deps.Seeder.CreateAppConfig(s.ctx, s.T(), s.testApp.ID)
	s.otherOrg = s.deps.Seeder.CreateOrg(s.ctx, s.T())

	s.svc = New(Params{
		V:               s.deps.V,
		DB:              s.deps.DB,
		L:               s.deps.L,
		InstallsHelpers: s.deps.InstallsHelpers,
	})

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.deps.L,
		DB:      s.deps.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	// The handler alone: require.Route has its own verb tests, and the runner
	// engine's chain is what puts org and account in context here.
	s.router.POST("/v1/stacks/:install_id/phone-home", s.svc.PostStackPhoneHome)
}

func (s *StackPhoneHomeTestSuite) post(installID string, body any) *httptest.ResponseRecorder {
	raw, err := json.Marshal(body)
	require.NoError(s.T(), err)

	req, err := http.NewRequest(http.MethodPost, "/v1/stacks/"+installID+"/phone-home", bytes.NewBuffer(raw))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	s.router.ServeHTTP(rr, req)
	return rr
}

// seedInstall returns an install in the given org with an install stack.
func (s *StackPhoneHomeTestSuite) seedInstall(orgID string) *app.Install {
	install := s.deps.Seeder.CreateInstall(s.ctx, s.T(), s.testApp)
	if orgID != install.OrgID {
		require.NoError(s.T(), s.deps.DB.Model(&app.Install{}).
			Where("id = ?", install.ID).Update("org_id", orgID).Error)
		install.OrgID = orgID
	}
	return install
}

// seedQueue creates the install-signals queue the enqueue looks up by owner+name.
func (s *StackPhoneHomeTestSuite) seedQueue(installID string) {
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&app.Queue{
		OwnerID:   installID,
		OwnerType: "installs",
		Name:      installshelpers.InstallSignalsQueueName,
	}).Error)
}

func (s *StackPhoneHomeTestSuite) TestAppliesReportToLatestStackVersion() {
	t := s.T()

	install := s.seedInstall(s.testOrg.ID)
	s.seedQueue(install.ID)

	// Two versions: the report belongs to the newest, the one being applied.
	older := s.deps.Seeder.CreateInstallStackVersion(s.ctx, t, install.ID, install.InstallStack.ID, s.appCfg.ID)
	latest := s.deps.Seeder.CreateInstallStackVersion(s.ctx, t, install.ID, install.InstallStack.ID, s.appCfg.ID)

	rr := s.post(install.ID, map[string]any{
		"request_type": installshelpers.PhoneHomeRequestTypeCreate,
		"runner_role":  "arn:aws:iam::000000000000:role/runner",
	})
	require.Equal(t, http.StatusCreated, rr.Code, "body: %s", rr.Body.String())

	var runs []app.InstallStackVersionRun
	require.NoError(t, s.deps.DB.
		Where("install_stack_version_id = ?", latest.ID).
		Find(&runs).Error)
	require.Len(t, runs, 1)
	require.NotNil(t, runs[0].Data["runner_role"])
	assert.Equal(t, "arn:aws:iam::000000000000:role/runner", *runs[0].Data["runner_role"])

	var reloaded app.InstallStackVersion
	require.NoError(t, s.deps.DB.Where("id = ?", latest.ID).First(&reloaded).Error)
	assert.Equal(t, app.InstallStackVersionStatusActive, reloaded.Status.Status)

	var olderRuns []app.InstallStackVersionRun
	require.NoError(t, s.deps.DB.
		Where("install_stack_version_id = ?", older.ID).
		Find(&olderRuns).Error)
	assert.Empty(t, olderRuns, "an older version must not receive the report")

	assert.NotEmpty(t, tests.GetQueueSignals(t, s.deps.DB), "the report must signal the install")
}

// The org scope is what keeps one org's stack credential from reporting into another
// org's install, so the miss must look identical to a nonexistent install.
func (s *StackPhoneHomeTestSuite) TestInstallInAnotherOrgIsNotFound() {
	t := s.T()

	install := s.seedInstall(s.otherOrg.ID)
	s.seedQueue(install.ID)
	s.deps.Seeder.CreateInstallStackVersion(s.ctx, t, install.ID, install.InstallStack.ID, s.appCfg.ID)

	rr := s.post(install.ID, map[string]any{
		"request_type": installshelpers.PhoneHomeRequestTypeCreate,
	})
	assert.Equal(t, http.StatusNotFound, rr.Code, "body: %s", rr.Body.String())

	rr = s.post("inst_does_not_exist_000000", map[string]any{
		"request_type": installshelpers.PhoneHomeRequestTypeCreate,
	})
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

// The config read tolerates a missing version — the module fetches config before the
// version exists. A report cannot: there is nothing to record it against.
func (s *StackPhoneHomeTestSuite) TestNoStackVersionIsNotFound() {
	t := s.T()

	install := s.seedInstall(s.testOrg.ID)

	rr := s.post(install.ID, map[string]any{
		"request_type": installshelpers.PhoneHomeRequestTypeCreate,
	})
	assert.Equal(t, http.StatusNotFound, rr.Code, "body: %s", rr.Body.String())
}

// Delete is accepted and dropped, matching the legacy route: a deprovisioned stack
// must stay deletable, and the report carries nothing worth recording.
func (s *StackPhoneHomeTestSuite) TestDeleteIsAcceptedWithoutRecording() {
	t := s.T()

	install := s.seedInstall(s.testOrg.ID)
	version := s.deps.Seeder.CreateInstallStackVersion(s.ctx, t, install.ID, install.InstallStack.ID, s.appCfg.ID)

	rr := s.post(install.ID, map[string]any{
		"request_type": installshelpers.PhoneHomeRequestTypeDelete,
	})
	require.Equal(t, http.StatusOK, rr.Code, "body: %s", rr.Body.String())

	var runs []app.InstallStackVersionRun
	require.NoError(t, s.deps.DB.
		Where("install_stack_version_id = ?", version.ID).
		Find(&runs).Error)
	assert.Empty(t, runs)
}

func (s *StackPhoneHomeTestSuite) TestRejectsBadRequestType() {
	t := s.T()

	install := s.seedInstall(s.testOrg.ID)

	assert.Equal(t, http.StatusBadRequest,
		s.post(install.ID, map[string]any{"request_type": "Destroy"}).Code)
	assert.Equal(t, http.StatusBadRequest,
		s.post(install.ID, map[string]any{"request_type": 7}).Code)
	assert.Equal(t, http.StatusBadRequest,
		s.post(install.ID, map[string]any{}).Code)
}
