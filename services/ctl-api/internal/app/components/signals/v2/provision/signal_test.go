package provision

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

	// Provision is a noop - should just succeed
	err := s.service.ExecuteSignal(ctx, signal)
	s.Require().NoError(err)
}
