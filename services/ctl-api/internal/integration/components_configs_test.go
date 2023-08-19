package integration

import (
	"os"
	"testing"

	"github.com/powertoolsdev/mono/pkg/api/client/models"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type componentConfigsSuite struct {
	baseIntegrationTestSuite

	orgID  string
	appID  string
	compID string
}

func TestComponentConfigsSuite(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	suite.Run(t, new(componentConfigsSuite))
}

func (s *componentConfigsSuite) SetupTest() {
	// create an org
	orgReq := generics.GetFakeObj[*models.ServiceCreateOrgRequest]()
	org, err := s.apiClient.CreateOrg(s.ctx, orgReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), org)
	s.apiClient.SetOrgID(org.ID)
	s.orgID = org.ID

	// add a vcs connection to the org
	vcsReq := generics.GetFakeObj[*models.ServiceCreateConnectionRequest]()
	_, err = s.apiClient.CreateVCSConnection(s.ctx, vcsReq)
	require.Nil(s.T(), err)

	// create an app
	appReq := generics.GetFakeObj[*models.ServiceCreateAppRequest]()
	app, err := s.apiClient.CreateApp(s.ctx, appReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), app)
	s.appID = app.ID

	// create a component
	compReq := generics.GetFakeObj[*models.ServiceCreateComponentRequest]()
	comp, err := s.apiClient.CreateComponent(s.ctx, s.appID, compReq)
	require.Nil(s.T(), err)
	require.NotNil(s.T(), comp)
	s.compID = comp.ID
}

func (s *componentConfigsSuite) TestCreateDockerBuildComponentConfig() {
	req := generics.GetFakeObj[*models.ServiceCreateDockerBuildComponentConfigRequest]()
	req.ConnectedGithubVcsConfig.Repo = generics.ToPtr("powertoolsdev/mono")

	s.T().Run("success with github config", func(t *testing.T) {
		cfg, err := s.apiClient.CreateDockerBuildComponentConfig(s.ctx, s.compID, req)
		require.Nil(t, err)
		require.NotNil(t, cfg)
	})

	s.T().Run("errors on invalid parameters", func(t *testing.T) {
	})

	s.T().Run("renders correctly as latest", func(t *testing.T) {
	})
}
