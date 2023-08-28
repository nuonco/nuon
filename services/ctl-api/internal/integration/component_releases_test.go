package integration

import (
	"os"
	"testing"

	"github.com/powertoolsdev/mono/pkg/api/client/models"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type componentReleasesTestSuite struct {
	baseIntegrationTestSuite

	orgID     string
	appID     string
	compID    string
	installID string
	buildID   string
}

func TestComponentReleasesSuite(t *testing.T) {
	t.Parallel()

	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	suite.Run(t, new(componentReleasesTestSuite))
}

func (s *componentReleasesTestSuite) TearDownTest() {
	s.deleteOrg(s.orgID)
}

func (s *componentReleasesTestSuite) SetupTest() {
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

	// create a component config
	req := generics.GetFakeObj[*models.ServiceCreateExternalImageComponentConfigRequest]()
	req.ConnectedGithubVcsConfig.Repo = generics.ToPtr("powertoolsdev/mono")
	cfg, err := s.apiClient.CreateExternalImageComponentConfig(s.ctx, s.compID, req)
	require.Nil(s.T(), err)
	require.NotNil(s.T(), cfg)

	// create a build of this component
	buildReq := &models.ServiceCreateComponentBuildRequest{
		GitRef: "HEAD",
	}
	build, err := s.apiClient.CreateComponentBuild(s.ctx, comp.ID, buildReq)
	require.NoError(s.T(), err)
	s.buildID = build.ID

	// create install
	fakeReq := generics.GetFakeObj[*models.ServiceCreateInstallRequest]()
	fakeReq.AwsAccount.Region = "us-west-2"
	install, err := s.apiClient.CreateInstall(s.ctx, s.appID, fakeReq)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), install)
	s.installID = install.ID
}

func (s *componentReleasesTestSuite) TestCreateRelease() {
	s.T().Run("success with parallel strategy", func(t *testing.T) {
		release, err := s.apiClient.CreateComponentRelease(s.ctx, s.compID, &models.ServiceCreateComponentReleaseRequest{
			BuildID: s.buildID,
			Strategy: &models.ServiceCreateComponentReleaseRequestStrategy{
				ReleaseStrategy: "parallel",
				InstallsPerStep: 1,
			},
		})
		require.NoError(t, err)
		require.NotEmpty(t, release)
	})

	s.T().Run("fails with missing component", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateComponentReleaseRequest]()
		release, err := s.apiClient.CreateComponentRelease(s.ctx, generics.GetFakeObj[string](), req)
		require.Error(t, err)
		require.Nil(t, release)
	})

	s.T().Run("fails with invalid build id", func(t *testing.T) {
		req := generics.GetFakeObj[*models.ServiceCreateComponentReleaseRequest]()
		release, err := s.apiClient.CreateComponentRelease(s.ctx, s.compID, req)
		require.Error(t, err)
		require.Nil(t, release)
	})
}

func (s *componentReleasesTestSuite) TestGetAppReleases() {
	release, err := s.apiClient.CreateComponentRelease(s.ctx, s.compID, &models.ServiceCreateComponentReleaseRequest{
		BuildID: s.buildID,
		Strategy: &models.ServiceCreateComponentReleaseRequestStrategy{
			ReleaseStrategy: "parallel",
			InstallsPerStep: 1,
		},
	})
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), release)

	s.T().Run("successfully returns from one component", func(t *testing.T) {
		releases, err := s.apiClient.GetAppReleases(s.ctx, s.appID)
		require.NoError(t, err)
		require.NotEmpty(t, releases)

		require.Equal(t, release.ID, releases[0].ID)
	})

	s.T().Run("returns them in the correct order", func(t *testing.T) {
		secondRelease, err := s.apiClient.CreateComponentRelease(s.ctx, s.compID, &models.ServiceCreateComponentReleaseRequest{
			BuildID: s.buildID,
			Strategy: &models.ServiceCreateComponentReleaseRequestStrategy{
				ReleaseStrategy: "parallel",
				InstallsPerStep: 1,
			},
		})
		require.NoError(t, err)
		require.NotEmpty(t, release)

		releases, err := s.apiClient.GetAppReleases(s.ctx, s.appID)
		require.NoError(s.T(), err)
		require.NotEmpty(t, releases)
		require.Len(t, releases, 2)

		require.Equal(t, release.ID, releases[1].ID)
		require.Equal(t, secondRelease.ID, releases[0].ID)
	})
}

func (s *componentReleasesTestSuite) TestGetComponentReleases() {
	release, err := s.apiClient.CreateComponentRelease(s.ctx, s.compID, &models.ServiceCreateComponentReleaseRequest{
		BuildID: s.buildID,
		Strategy: &models.ServiceCreateComponentReleaseRequestStrategy{
			ReleaseStrategy: "parallel",
			InstallsPerStep: 1,
		},
	})
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), release)

	s.T().Run("successfully returns from component", func(t *testing.T) {
		releases, err := s.apiClient.GetComponentReleases(s.ctx, s.compID)
		require.NoError(t, err)
		require.NotEmpty(t, releases)

		require.Equal(t, release.ID, releases[0].ID)
	})

	s.T().Run("returns in desc created at order", func(t *testing.T) {
		secondRelease, err := s.apiClient.CreateComponentRelease(s.ctx, s.compID, &models.ServiceCreateComponentReleaseRequest{
			BuildID: s.buildID,
			Strategy: &models.ServiceCreateComponentReleaseRequestStrategy{
				ReleaseStrategy: "parallel",
				InstallsPerStep: 1,
			},
		})
		require.NoError(t, err)
		require.NotEmpty(t, release)

		releases, err := s.apiClient.GetComponentReleases(s.ctx, s.compID)
		require.NoError(t, err)
		require.NotEmpty(t, releases)
		require.Len(t, releases, 2)

		require.Equal(t, release.ID, releases[1].ID)
		require.Equal(t, secondRelease.ID, releases[0].ID)
	})
}

func (s *componentReleasesTestSuite) TestGetComponentRelease() {
	release, err := s.apiClient.CreateComponentRelease(s.ctx, s.compID, &models.ServiceCreateComponentReleaseRequest{
		BuildID: s.buildID,
		Strategy: &models.ServiceCreateComponentReleaseRequestStrategy{
			ReleaseStrategy: "parallel",
			InstallsPerStep: 1,
		},
	})
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), release)

	s.T().Run("successfully gets release by id", func(t *testing.T) {
		fetched, err := s.apiClient.GetRelease(s.ctx, release.ID)
		require.NoError(t, err)
		require.NotNil(t, fetched)
		require.Equal(t, release.ID, fetched.ID)
	})

	s.T().Run("fails when id is invalid", func(t *testing.T) {
		fetched, err := s.apiClient.GetRelease(s.ctx, generics.GetFakeObj[string]())
		require.Error(t, err)
		require.Empty(t, fetched)
	})
}

func (s *componentReleasesTestSuite) TestGetComponentReleaseSteps() {
	release, err := s.apiClient.CreateComponentRelease(s.ctx, s.compID, &models.ServiceCreateComponentReleaseRequest{
		BuildID: s.buildID,
		Strategy: &models.ServiceCreateComponentReleaseRequestStrategy{
			ReleaseStrategy: "parallel",
			InstallsPerStep: 1,
		},
	})
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), release)

	s.T().Run("successfully gets steps by release id", func(t *testing.T) {
		steps, err := s.apiClient.GetReleaseSteps(s.ctx, release.ID)
		require.NoError(t, err)
		require.Len(t, steps, 1)
	})

	s.T().Run("fails when id is invalid", func(t *testing.T) {
		fetched, err := s.apiClient.GetReleaseSteps(s.ctx, generics.GetFakeObj[string]())
		require.Error(t, err)
		require.Empty(t, fetched)
	})
}
