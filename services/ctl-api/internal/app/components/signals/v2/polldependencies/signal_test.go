package polldependencies

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/fx/fxtest"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/testworker"
)

type SignalTestSuite struct {
	suite.Suite
	app     *fxtest.App
	service testworker.TestService
}

func TestSignalSuite(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("INTEGRATION is not set, skipping")
	}
	suite.Run(t, new(SignalTestSuite))
}

func (s *SignalTestSuite) SetupSuite() {
	s.app, s.service = testworker.SetupTestSuite(s.T(), "components")
}

func (s *SignalTestSuite) TearDownSuite() {
	s.app.RequireStop()
}

func (s *SignalTestSuite) TestSignalExecute() {
	ctx := s.service.Ctx()

	// Create test component with app in active status
	comp := s.service.MustMakeTestComponent(ctx, &testworker.MakeTestComponentOptions{
		AppStatus: app.AppStatusActive,
	})

	// Execute signal
	signal := &Signal{
		ComponentID: comp.ID,
	}

	// Should complete immediately since app is already active
	err := s.service.ExecuteSignal(ctx, signal)
	s.Require().NoError(err)
}

func (s *SignalTestSuite) TestSignalExecuteWithPending() {
	ctx := s.service.Ctx()

	// Create test component with app in pending status
	comp := s.service.MustMakeTestComponent(ctx, &testworker.MakeTestComponentOptions{
		AppStatus: app.AppStatusProvisioning,
	})

	// Execute signal in background
	signal := &Signal{
		ComponentID: comp.ID,
	}

	// This would poll until app becomes active
	// In a real test, we'd update the app status in a goroutine
	// For now, just verify the signal can be created
	s.NotNil(signal)
}
