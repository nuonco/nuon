package configcreated

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/fx/fxtest"

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

	// Create test component
	comp := s.service.MustMakeTestComponent(ctx, nil)

	// Execute signal
	signal := &Signal{
		ComponentID: comp.ID,
	}

	err := s.service.ExecuteSignal(ctx, signal)
	s.Require().NoError(err)

	// Verify a build was queued (check that a component build exists)
	builds, err := s.service.GetComponentBuilds(ctx, comp.ID)
	s.Require().NoError(err)
	s.Greater(len(builds), 0, "expected at least one build to be queued")
}
