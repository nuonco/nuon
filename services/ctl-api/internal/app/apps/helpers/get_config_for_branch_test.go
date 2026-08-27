package helpers_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/stretchr/testify/suite"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

type getConfigForBranchDeps struct {
	fx.In

	DB      *gorm.DB `name:"psql"`
	Seed    *testseed.Seeder
	Helpers *appshelpers.Helpers
}

type GetLatestActiveAppConfigForBranchTestSuite struct {
	tests.BaseDBTestSuite

	app  *fxtest.App
	deps getConfigForBranchDeps
}

func TestGetLatestActiveAppConfigForBranchSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(GetLatestActiveAppConfigForBranchTestSuite))
}

func (s *GetLatestActiveAppConfigForBranchTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()

	options := append(tests.CtlApiFXOptions(s.T()), fx.Populate(&s.deps))
	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.deps.DB)
}

func (s *GetLatestActiveAppConfigForBranchTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *GetLatestActiveAppConfigForBranchTestSuite) seedBranch(ctx context.Context, appID, name string) *app.AppBranch {
	branch := &app.AppBranch{AppID: appID, Name: name}
	s.Require().NoError(s.deps.DB.WithContext(ctx).Create(branch).Error)
	return branch
}

func (s *GetLatestActiveAppConfigForBranchTestSuite) seedConfig(
	ctx context.Context,
	appID, branchID string,
	createdAt time.Time,
	componentIDs []string,
) *app.AppConfig {
	cfg := &app.AppConfig{
		AppID:        appID,
		CreatedAt:    createdAt,
		Status:       app.AppConfigStatusActive,
		StatusV2:     app.NewCompositeStatus(ctx, app.Status(app.AppConfigStatusActive)),
		CLIVersion:   "development",
		ComponentIDs: pq.StringArray(componentIDs),
	}
	if branchID != "" {
		s.Require().NoError(cfg.AppBranchID.Scan(branchID))
	}
	s.Require().NoError(s.deps.DB.WithContext(ctx).Create(cfg).Error)
	return cfg
}

func (s *GetLatestActiveAppConfigForBranchTestSuite) TestReturnsBranchConfigOverNewerConfigFromAnotherBranch() {
	ctx := context.Background()
	ctx, _ = s.deps.Seed.EnsureAccount(ctx, s.T())
	ctx, _ = s.deps.Seed.EnsureOrg(ctx, s.T())

	testApp := s.deps.Seed.CreateApp(ctx, s.T())
	now := time.Now().UTC()

	mine := s.seedBranch(ctx, testApp.ID, "mine")
	theirs := s.seedBranch(ctx, testApp.ID, "theirs")

	want := s.seedConfig(ctx, testApp.ID, mine.ID, now.Add(-2*time.Hour), []string{"cmp-mine"})
	s.seedConfig(ctx, testApp.ID, theirs.ID, now, []string{"cmp-theirs"})

	got, err := s.deps.Helpers.GetLatestActiveAppConfigForBranch(ctx, testApp.ID, mine.ID)
	s.Require().NoError(err)
	s.Equal(want.ID, got.ID)
	s.Equal([]string{"cmp-mine"}, []string(got.ComponentIDs))
}

func (s *GetLatestActiveAppConfigForBranchTestSuite) TestFallsBackToAppConfigWhenBranchHasNeverSynced() {
	ctx := context.Background()
	ctx, _ = s.deps.Seed.EnsureAccount(ctx, s.T())
	ctx, _ = s.deps.Seed.EnsureOrg(ctx, s.T())

	testApp := s.deps.Seed.CreateApp(ctx, s.T())
	now := time.Now().UTC()

	neverRun := s.seedBranch(ctx, testApp.ID, "never-run")
	s.seedConfig(ctx, testApp.ID, "", now.Add(-2*time.Hour), []string{"cmp-old"})
	want := s.seedConfig(ctx, testApp.ID, "", now, []string{"cmp-new"})

	got, err := s.deps.Helpers.GetLatestActiveAppConfigForBranch(ctx, testApp.ID, neverRun.ID)
	s.Require().NoError(err)
	s.Equal(want.ID, got.ID)
	s.Equal([]string{"cmp-new"}, []string(got.ComponentIDs))
}

func (s *GetLatestActiveAppConfigForBranchTestSuite) TestErrorsWhenBranchBelongsToAnotherApp() {
	ctx := context.Background()
	ctx, _ = s.deps.Seed.EnsureAccount(ctx, s.T())
	ctx, _ = s.deps.Seed.EnsureOrg(ctx, s.T())

	appA := s.deps.Seed.CreateApp(ctx, s.T())
	appB := s.deps.Seed.CreateApp(ctx, s.T())
	branchB := s.seedBranch(ctx, appB.ID, "other")

	_, err := s.deps.Helpers.GetLatestActiveAppConfigForBranch(ctx, appA.ID, branchB.ID)
	s.Require().Error(err)
	s.ErrorIs(err, appshelpers.ErrAppBranchNotFound)
}
