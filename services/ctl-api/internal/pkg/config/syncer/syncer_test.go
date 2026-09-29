package syncer

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/tests"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
	testseedconfig "github.com/nuonco/nuon/services/ctl-api/tests/testseed/config"
)

type TestService struct {
	DB   *gorm.DB `name:"psql"`
	L    *zap.Logger
	Seed *testseed.Seeder
}

type SyncerTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	service TestService

	testAccount *app.Account
	testOrg     *app.Org
	testApp     *app.App
}

func TestSyncerSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(SyncerTestSuite))
}

func (s *SyncerTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()

	options := tests.CtlApiFXOptions(s.T())

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()
}

func (s *SyncerTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *SyncerTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()

	// TODO: Seed test data once testseed is integrated
	// ctx := context.Background()
	// ctx = s.service.Seed.EnsureAccount(ctx, s.T())
	// ctx = s.service.Seed.EnsureOrg(ctx, s.T())
	// s.testApp = s.service.Seed.EnsureApp(ctx, s.T())
}

func (s *SyncerTestSuite) TearDownTest() {
}

func (s *SyncerTestSuite) TestSmokeTest() {
	cfg := testseedconfig.BuildMinimalAppConfig()
	s.NotNil(cfg, "BuildMinimalAppConfig should return a config")
	s.Equal("1", cfg.Version, "Version should be set")
	s.NotNil(cfg.Sandbox, "Sandbox should be set")
	s.NotNil(cfg.Runner, "Runner should be set")
}

func (s *SyncerTestSuite) TestMinimalSync() {
	s.T().Skip("TODO: Implement after testseed integration")

	// TODO: Uncomment when testseed is integrated with FX
	/*
		ctx := context.Background()
		ctx = s.service.Seed.EnsureAccount(ctx, s.T())
		ctx = s.service.Seed.EnsureOrg(ctx, s.T())
		testApp := s.service.Seed.EnsureApp(ctx, s.T())

		// Create a minimal config
		cfg := testseedconfig.BuildMinimalAppConfig()

		// Create syncer
		syncerInstance := New(Params{DB: s.service.DB}, testApp.ID, cfg)

		// Execute sync
		err := syncerInstance.Sync(ctx)
		s.NoError(err, "Sync should succeed with minimal config")

		// Verify app config was created in database
		var appConfig app.AppConfig
		err = s.service.DB.Where("app_id = ?", testApp.ID).
			Order("created_at DESC").
			First(&appConfig).Error
		s.NoError(err, "Should find created app config")
		s.Equal(app.AppConfigStatusActive, appConfig.Status)
	*/
}
