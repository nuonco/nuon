package updatecomponenttype

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

	// Create test component with initial type
	comp := s.service.MustMakeTestComponent(ctx, &testworker.MakeTestComponentOptions{
		Type: app.ComponentTypeDockerBuild,
	})

	// Execute signal to change component type
	signal := &Signal{
		ComponentID:   comp.ID,
		ComponentType: app.ComponentTypeHelm,
	}

	err := s.service.ExecuteSignal(ctx, signal)
	s.Require().NoError(err)

	// Verify component type was updated
	updated, err := s.service.GetComponent(ctx, comp.ID)
	s.Require().NoError(err)
	s.Equal(app.ComponentTypeHelm, updated.Type)
}
