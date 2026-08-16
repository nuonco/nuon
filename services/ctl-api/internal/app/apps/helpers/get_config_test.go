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

type getConfigDeps struct {
	fx.In

	DB      *gorm.DB `name:"psql"`
	Seed    *testseed.Seeder
	Helpers *appshelpers.Helpers
}

type GetLatestActiveAppConfigTestSuite struct {
	tests.BaseDBTestSuite

	app  *fxtest.App
	deps getConfigDeps
}

func TestGetLatestActiveAppConfigSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}
	suite.Run(t, new(GetLatestActiveAppConfigTestSuite))
}

func (s *GetLatestActiveAppConfigTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()

	options := append(tests.CtlApiFXOptions(s.T()), fx.Populate(&s.deps))
	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
	s.SetDB(s.deps.DB)
}

func (s *GetLatestActiveAppConfigTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

// seedAppConfig persists a config with an explicit ID, created_at and status so the
// test controls the ordering dimensions independently.
func (s *GetLatestActiveAppConfigTestSuite) seedAppConfig(ctx context.Context, id, appID string, createdAt time.Time, status app.AppConfigStatus) *app.AppConfig {
	cfg := &app.AppConfig{
		ID:         id,
		AppID:      appID,
		CreatedAt:  createdAt,
		Status:     status,
		StatusV2:   app.NewCompositeStatus(ctx, app.Status(status)),
		CLIVersion: "development",
	}
	s.Require().NoError(s.deps.DB.WithContext(ctx).Create(cfg).Error)
	return cfg
}

// Ensure we return only the newest active config, not simply the smallest, or the newest.
// This is to guard against regressions in the ordering of app config version.
func (s *GetLatestActiveAppConfigTestSuite) TestReturnsNewestActiveConfigNotSmallestID() {
	ctx := context.Background()
	ctx, _ = s.deps.Seed.EnsureAccount(ctx, s.T())
	ctx, _ = s.deps.Seed.EnsureOrg(ctx, s.T())

	testApp := s.deps.Seed.CreateApp(ctx, s.T())
	now := time.Now().UTC()

	oldest := s.seedAppConfig(ctx, "app00000000000000000000000", testApp.ID, now.Add(-2*time.Hour), app.AppConfigStatusActive)
	newest := s.seedAppConfig(ctx, "appzzzzzzzzzzzzzzzzzzzzzz0", testApp.ID, now.Add(-1*time.Hour), app.AppConfigStatusActive)
	// A newer non-active config must never win: the status filter still applies.
	s.seedAppConfig(ctx, "appzzzzzzzzzzzzzzzzzzzzzz1", testApp.ID, now, app.AppConfigStatusError)

	got, err := s.deps.Helpers.GetLatestActiveAppConfig(ctx, testApp.ID)
	s.Require().NoError(err)
	s.Equal(newest.ID, got.ID, "expected the newest active config, not %s (oldest/smallest ID)", oldest.ID)

	gotBare, err := s.deps.Helpers.GetLatestActiveAppConfigBare(ctx, testApp.ID)
	s.Require().NoError(err)
	s.Equal(newest.ID, gotBare.ID, "expected the newest active config, not %s (oldest/smallest ID)", oldest.ID)
}
