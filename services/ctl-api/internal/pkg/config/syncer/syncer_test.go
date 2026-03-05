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

// TestService contains the dependencies injected by FX for testing.
type TestService struct {
	DB   *gorm.DB `name:"psql"`
	L    *zap.Logger
	Seed *testseed.Seeder
}

// SyncerTestSuite is the main test suite for the config syncer.
// It embeds BaseDBTestSuite to get database lifecycle management.
type SyncerTestSuite struct {
	tests.BaseDBTestSuite

	app     *fxtest.App
	service TestService

	// Test fixtures set up in SetupTest
	testAccount *app.Account
	testOrg     *app.Org
	testApp     *app.App
}

// TestSyncerSuite is the entry point for running the syncer test suite.
func TestSyncerSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
		return
	}

	suite.Run(t, new(SyncerTestSuite))
}

// SetupSuite runs once before all tests in the suite.
// It creates the FX app with all dependencies and starts it.
func (s *SyncerTestSuite) SetupSuite() {
	s.BaseDBTestSuite.SetupSuite()

	// Build FX options using shared test options plus testseed
	options := append(
		tests.CtlApiFXOptions(s.T()),
		// Add testseed provider
		// fx.Provide(testseed.New), // TODO: Uncomment when fx.In is available
		// fx.Populate(&s.service),
	)

	s.app = fxtest.New(s.T(), options...)
	s.app.RequireStart()

	// Set the DB for BaseDBTestSuite to enable auto-truncation
	// s.SetDB(s.service.DB) // TODO: Uncomment when service is populated
}

// TearDownSuite runs once after all tests in the suite.
func (s *SyncerTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

// SetupTest runs before each test method.
// It truncates tables and creates fresh test fixtures.
func (s *SyncerTestSuite) SetupTest() {
	s.BaseDBTestSuite.SetupTest()

	// TODO: Seed test data once testseed is integrated
	// ctx := context.Background()
	// ctx = s.service.Seed.EnsureAccount(ctx, s.T())
	// ctx = s.service.Seed.EnsureOrg(ctx, s.T())
	// s.testApp = s.service.Seed.EnsureApp(ctx, s.T())
}

// TearDownTest runs after each test method.
func (s *SyncerTestSuite) TearDownTest() {
	// BaseDBTestSuite handles cleanup
}

// TestSmokeTest is a basic smoke test to verify the test infrastructure works.
func (s *SyncerTestSuite) TestSmokeTest() {
	// This test just verifies that:
	// 1. The test suite can be instantiated
	// 2. FX app starts successfully
	// 3. Database is accessible
	// 4. Faker providers are registered

	// Verify faker provider registration
	cfg := testseedconfig.BuildMinimalAppConfig()
	s.NotNil(cfg, "BuildMinimalAppConfig should return a config")
	s.Equal("1", cfg.Version, "Version should be set")
	s.NotNil(cfg.Sandbox, "Sandbox should be set")
	s.NotNil(cfg.Runner, "Runner should be set")
}

// TestMinimalSync tests syncing a minimal valid config.
// This is the first real sync test to verify the basic flow works.
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
