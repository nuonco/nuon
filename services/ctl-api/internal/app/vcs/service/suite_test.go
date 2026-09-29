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
	"github.com/google/go-github/v50/github"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type VCSServiceTestDeps struct {
	fx.In

	DB     *gorm.DB `name:"psql"`
	CHDB   *gorm.DB `name:"ch"`
	V      *validator.Validate
	L      *zap.Logger
	Seeder *testseed.Seeder
}

type VCSServiceTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	service VCSServiceTestDeps
	router  *gin.Engine
	ctx     context.Context
	testOrg *app.Org
	testAcc *app.Account
	ctrl    *gomock.Controller
	mockGH  *vcshelpers.MockGithubClient
}

func TestVCSServiceSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(VCSServiceTestSuite))
}

func (s *VCSServiceTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()
	gin.SetMode(gin.TestMode)

	s.ctrl = gomock.NewController(s.T())
	s.mockGH = vcshelpers.NewMockGithubClient(s.ctrl)

	options := append(
		tests.CtlApiFXOptionsWithMocks(tests.TestOpts{
			T: s.T(),
			Mocks: &tests.TestMocks{
				MockGH: s.mockGH,
			},
			CustomValidator: true,
		}),
		fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()

	s.SetDB(s.service.DB)
}

func (s *VCSServiceTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()
	s.setupTestData()

	s.mockGH.EXPECT().GetInstallationAccount(gomock.Any(), gomock.Any()).Return(&github.User{
		Login: github.String("test-org"),
		ID:    github.Int64(12345),
		Type:  github.String("Organization"),
	}, nil).AnyTimes()

	s.mockGH.EXPECT().GetInstallation(gomock.Any(), gomock.Any()).Return(&github.Installation{
		ID: github.Int64(12345),
		Account: &github.User{
			Login: github.String("test-org"),
			ID:    github.Int64(67890),
			Type:  github.String("Organization"),
		},
		Permissions: &github.InstallationPermissions{
			Contents: github.String("read"),
			Metadata: github.String("read"),
		},
		RepositorySelection: github.String("all"),
	}, nil).AnyTimes()

	s.mockGH.EXPECT().DeleteInstallation(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	s.mockGH.EXPECT().ListInstallationRepos(gomock.Any(), gomock.Any()).Return([]*github.Repository{
		{
			ID:            github.Int64(1),
			Name:          github.String("test-repo"),
			FullName:      github.String("test-org/test-repo"),
			Private:       github.Bool(false),
			DefaultBranch: github.String("main"),
		},
	}, nil).AnyTimes()

	svc := &service{
		l:        s.service.L,
		db:       s.service.DB,
		v:        s.service.V,
		ghClient: s.mockGH,
	}

	s.router = tests.NewTestRouter(tests.RouterOptions{
		L:       s.service.L,
		DB:      s.service.DB,
		TestOrg: s.testOrg,
		TestAcc: s.testAcc,
	})

	err := svc.RegisterPublicRoutes(s.router)
	require.NoError(s.T(), err)
}

func (s *VCSServiceTestSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *VCSServiceTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *VCSServiceTestSuite) setupTestData() {
	s.ctx = context.Background()
	s.ctx, s.testAcc = s.service.Seeder.EnsureAccount(s.ctx, s.T())
	s.ctx, s.testOrg = s.service.Seeder.EnsureOrg(s.ctx, s.T())
}

func (s *VCSServiceTestSuite) createTestVCSConnection() *app.VCSConnection {
	conn := &app.VCSConnection{
		OrgID:             s.testOrg.ID,
		GithubInstallID:   "12345",
		GithubAccountID:   "67890",
		GithubAccountName: "test-org",
	}
	err := s.service.DB.WithContext(s.ctx).Create(conn).Error
	require.NoError(s.T(), err)
	return conn
}

func (s *VCSServiceTestSuite) makeRequest(method, path string, body interface{}) *httptest.ResponseRecorder {
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
