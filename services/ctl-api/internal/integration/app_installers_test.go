package integration

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type appInstallersSuite struct {
	baseIntegrationTestSuite

	orgID string
	appID string
}

func TestAppInstallersSuite(t *testing.T) {
	t.Parallel()

	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	suite.Run(t, new(appInstallersSuite))
}

func (s *appInstallersSuite) TearDownTest() {
	s.deleteOrg(s.orgID)
}

func (s *appInstallersSuite) SetupTest() {
	// create an org
	org := s.createOrg()
	s.orgID = org.ID

	app := s.createApp()
	s.appID = app.ID
}

func (s *appInstallersSuite) TestCreateAppInstaller() {
	s.T().Skip("todo - app installer create tests")
}

func (s *appInstallersSuite) TestGetAppInstaller() {
	s.T().Skip("todo - app installer get tests")
}

func (s *appInstallersSuite) TestUpdateAppInstaller() {
	s.T().Skip("todo - app installer update tests")
}

func (s *appInstallersSuite) TestDeleteAppInstaller() {
	s.T().Skip("todo - app installer delete tests")
}

func (s *appInstallersSuite) TestAppInstallerCreateInstall() {
	s.T().Skip("todo - app installer create install tests")
}

func (s *appInstallersSuite) TestAppInstallerGetInstall() {
	s.T().Skip("todo - app installer get install tests")
}
