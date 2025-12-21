package build

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

	// Create test component with active status
	comp := s.service.MustMakeTestComponent(ctx, &testworker.MakeTestComponentOptions{
		Status: app.ComponentStatusActive,
	})

	// Create a component build
	build := s.service.MustMakeTestComponentBuild(ctx, comp.ID)

	// Execute signal
	signal := &Signal{
		ComponentID: comp.ID,
		BuildID:     build.ID,
		SandboxMode: true,
	}

	err := s.service.ExecuteSignal(ctx, signal)
	s.Require().NoError(err)

	// Verify build status was updated to active
	updatedBuild, err := s.service.GetComponentBuild(ctx, build.ID)
	s.Require().NoError(err)
	s.Equal(app.ComponentBuildStatusActive, updatedBuild.Status)
}

func (s *SignalTestSuite) TestSignalExecuteInactiveComponent() {
	ctx := s.service.Ctx()

	// Create test component with inactive status
	comp := s.service.MustMakeTestComponent(ctx, &testworker.MakeTestComponentOptions{
		Status: app.ComponentStatusProvisioning,
	})

	// Create a component build
	build := s.service.MustMakeTestComponentBuild(ctx, comp.ID)

	// Execute signal
	signal := &Signal{
		ComponentID: comp.ID,
		BuildID:     build.ID,
		SandboxMode: true,
	}

	err := s.service.ExecuteSignal(ctx, signal)
	s.Require().Error(err)
	s.Contains(err.Error(), "component is not active")

	// Verify build status was updated to error
	updatedBuild, err := s.service.GetComponentBuild(ctx, build.ID)
	s.Require().NoError(err)
	s.Equal(app.ComponentBuildStatusError, updatedBuild.Status)
}
