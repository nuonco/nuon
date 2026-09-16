package helpers_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type findBaseAppBranchRunDeps struct {
	fx.In

	DB      *gorm.DB `name:"psql"`
	Seed    *testseed.Seeder
	Helpers *appshelpers.Helpers
}

type FindBaseAppBranchRunTestSuite struct {
	tests.BaseDBTestSuite

	app  *fxtest.App
	deps findBaseAppBranchRunDeps
}

func TestFindBaseAppBranchRunSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(FindBaseAppBranchRunTestSuite))
}

func (s *FindBaseAppBranchRunTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()

	options := append(tests.CtlApiFXOptions(s.T()), fx.Populate(&s.deps))
	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.deps.DB)
}

func (s *FindBaseAppBranchRunTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *FindBaseAppBranchRunTestSuite) TestPreviewUsesLatestRunOnTargetBranch() {
	ctx := context.Background()
	ctx, _ = s.deps.Seed.EnsureAccount(ctx, s.T())
	ctx, _ = s.deps.Seed.EnsureOrg(ctx, s.T())
	testApp := s.deps.Seed.CreateApp(ctx, s.T())

	mainBranch := &app.AppBranch{AppID: testApp.ID, Name: "main"}
	s.Require().NoError(s.deps.DB.WithContext(ctx).Create(mainBranch).Error)
	previewBranch := &app.AppBranch{AppID: testApp.ID, Name: "preview-owner"}
	s.Require().NoError(s.deps.DB.WithContext(ctx).Create(previewBranch).Error)

	mainConfig := &app.AppBranchConfig{AppBranchID: mainBranch.ID}
	s.Require().NoError(s.deps.DB.WithContext(ctx).Create(mainConfig).Error)
	previewConfig := &app.AppBranchConfig{AppBranchID: previewBranch.ID}
	s.Require().NoError(s.deps.DB.WithContext(ctx).Create(previewConfig).Error)

	completedAt := time.Now().UTC().Add(-time.Minute)
	mainRun := &app.AppBranchRun{
		AppBranchID:       mainBranch.ID,
		AppBranchConfigID: mainConfig.ID,
		RunType:           app.AppBranchRunTypeGit,
		Status:            "success",
		CompletedAt:       &completedAt,
	}
	s.Require().NoError(s.deps.DB.WithContext(ctx).Create(mainRun).Error)

	headRun := &app.AppBranchRun{
		AppBranchID:       previewBranch.ID,
		AppBranchConfigID: previewConfig.ID,
		RunType:           app.AppBranchRunTypeGitPreview,
		Status:            "pending",
		BaseBranch:        "main",
	}
	s.Require().NoError(s.deps.DB.WithContext(ctx).Create(headRun).Error)

	baseRun, err := s.deps.Helpers.FindBaseAppBranchRunForHead(ctx, headRun)
	s.Require().NoError(err)
	s.Equal(mainRun.ID, baseRun.ID)
}
