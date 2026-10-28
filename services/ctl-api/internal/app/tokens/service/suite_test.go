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
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	accountshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type TokensServiceTestDeps struct {
	fx.In

	DB              *gorm.DB `name:"psql"`
	L               *zap.Logger
	AcctClient      *account.Client
	AuthzClient     *authz.Client
	AccountsHelpers *accountshelpers.Helpers
	TokensService   *service
	Seeder          *testseed.Seeder
}

type TokensServiceTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	service TokensServiceTestDeps
	router  *gin.Engine
	ctx     context.Context
	testOrg *app.Org
	testAcc *app.Account
}

func TestTokensServiceSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(TokensServiceTestSuite))
}

func (s *TokensServiceTestSuite) SetupSuite() {
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

func (s *TokensServiceTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	require.NoError(s.T(), s.service.TokensService.RegisterPublicRoutes(s.router))
}

func (s *TokensServiceTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *TokensServiceTestSuite) setupTestData() {
	s.ctx = context.Background()
	s.ctx, s.testAcc = s.service.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.service.Seeder.EnsureOrg(s.ctx, s.T())
}

func (s *TokensServiceTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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
