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
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type OIDCFederationTestDeps struct {
	fx.In

	DB          *gorm.DB `name:"psql"`
	V           *validator.Validate
	L           *zap.Logger
	MW          metrics.Writer
	AuthzClient *authz.Client
	Service     *service
	Seeder      *testseed.Seeder
}

type OIDCFederationTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	deps    OIDCFederationTestDeps
	router  *gin.Engine
	ctx     context.Context
	testOrg *app.Org
	testAcc *app.Account
}

func TestOIDCFederationSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(OIDCFederationTestSuite))
}

func (s *OIDCFederationTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	options := append(
		tests.CtlApiFXOptions(s.T()),
		fx.Provide(New),
		fx.Populate(&s.deps),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()

	s.SetDB(s.deps.DB)
}

func (s *OIDCFederationTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.deps.Service.cfg.OIDCFederationEnabled = true
	s.setupTestData()

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.deps.L,
		DB:      s.deps.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := s.deps.Service.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *OIDCFederationTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

// setupTestData seeds an account and org, creates the org's standard roles,
// and makes the account an org admin so trust-policy CRUD is permitted.
func (s *OIDCFederationTestSuite) setupTestData() {
	s.ctx = context.Background()
	s.ctx, s.testAcc = s.deps.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.deps.Seeder.EnsureOrg(s.ctx, s.T())

	require.NoError(s.T(), s.deps.AuthzClient.CreateOrgRoles(s.ctx, s.testOrg.ID))
	require.NoError(s.T(), s.deps.AuthzClient.AddAccountOrgRole(s.ctx, app.RoleTypeOrgAdmin, s.testOrg.ID, s.testAcc.ID))

	var acct app.Account
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Preload("Roles").
		Preload("Roles.Org").
		Preload("Roles.Policies").
		Where("id = ?", s.testAcc.ID).
		First(&acct).Error)
	s.testAcc = &acct
}

// demoteTestAccount rebuilds the router with an account that has no org roles.
func (s *OIDCFederationTestSuite) demoteTestAccount() {
	_, nonAdmin := s.deps.Seeder.EnsureAccount(context.Background(), s.T())

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.deps.L,
		DB:      s.deps.DB,
		TestOrg: s.testOrg,
		TestAcc: nonAdmin,
	})
	require.NoError(s.T(), s.deps.Service.RegisterPublicRoutes(s.router))
}

func (s *OIDCFederationTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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
