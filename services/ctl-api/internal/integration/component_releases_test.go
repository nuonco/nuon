package integration

import (
	"os"
	"testing"
	"time"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type releasesTestSuite struct {
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

	suite.Run(t, new(releasesTestSuite))
}

func (s *releasesTestSuite) TearDownTest() {
	s.deleteOrg(s.orgID)
}

func (s *releasesTestSuite) SetupTest() {
	// create an org
	org := s.createOrg()
	s.orgID = org.ID

	app := s.createApp(s.orgID)
	s.appID = app.ID

	// create a component
	compReq := generics.GetFakeObj[*models.ServiceCreateComponentRequest]()
	comp, err := s.apiClient.CreateComponent(s.ctx, s.appID, compReq)
	require.Nil(s.T(), err)
	require.NotNil(s.T(), comp)
	s.compID = comp.ID

	// create a component config
	req := generics.GetFakeObj[*models.ServiceCreateExternalImageComponentConfigRequest]()
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
	install := s.createInstall(s.appID)
	s.installID = install.ID
}

func (s *releasesTestSuite) TestCreateRelease() {
	s.T().Run("success with parallel deploys", func(t *testing.T) {
		release, err := s.apiClient.CreateComponentRelease(s.ctx, s.compID, &models.ServiceCreateComponentReleaseRequest{
			BuildID: s.buildID,
			Strategy: &models.ServiceCreateComponentReleaseRequestStrategy{
				InstallsPerStep: 0,
				Delay:           generics.GetFakeObj[time.Duration]().String(),
			},
		})
		require.NoError(t, err)
		require.NotEmpty(t, release)
	})

	s.T().Run("success with 10 deploys at a time", func(t *testing.T) {
		release, err := s.apiClient.CreateComponentRelease(s.ctx, s.compID, &models.ServiceCreateComponentReleaseRequest{
			BuildID: s.buildID,
			Strategy: &models.ServiceCreateComponentReleaseRequestStrategy{
				InstallsPerStep: 10,
				Delay:           generics.GetFakeObj[time.Duration]().String(),
			},
		})
		require.NoError(t, err)
		require.NotEmpty(t, release)
	})

	s.T().Run("success with 10 deploys at a time and a 1 minute delay", func(t *testing.T) {
		release, err := s.apiClient.CreateComponentRelease(s.ctx, s.compID, &models.ServiceCreateComponentReleaseRequest{
			BuildID: s.buildID,
			Strategy: &models.ServiceCreateComponentReleaseRequestStrategy{
				InstallsPerStep: 10,
				Delay:           "1m",
			},
		})
		require.NoError(t, err)
		require.NotEmpty(t, release)
	})

	// s.T().Run("success with only a build ID", func(t *testing.T) {
	//	release, err := s.apiClient.CreateRelease(s.ctx, s.compID, &models.ServiceCreateReleaseRequest{
	//		BuildID: s.buildID,
	//		Strategy: &models.ServiceCreateComponentReleaseRequestStrategy{
	//			InstallsPerStep: 10,
	//			Delay:		 "1m",
	//		},
	//	})
	//	require.NoError(t, err)
	//	require.NotEmpty(t, release)
	// })

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

func (s *releasesTestSuite) TestGetAppReleases() {
	release, err := s.apiClient.CreateComponentRelease(s.ctx, s.compID, &models.ServiceCreateComponentReleaseRequest{
		BuildID: s.buildID,
		Strategy: &models.ServiceCreateComponentReleaseRequestStrategy{
			InstallsPerStep: 1,
			Delay:           generics.GetFakeObj[time.Duration]().String(),
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
				InstallsPerStep: 1,
				Delay:           generics.GetFakeObj[time.Duration]().String(),
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

func (s *releasesTestSuite) TestGetComponentReleases() {
	release, err := s.apiClient.CreateComponentRelease(s.ctx, s.compID, &models.ServiceCreateComponentReleaseRequest{
		BuildID: s.buildID,
		Strategy: &models.ServiceCreateComponentReleaseRequestStrategy{
			InstallsPerStep: 1,
			Delay:           generics.GetFakeObj[time.Duration]().String(),
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
				InstallsPerStep: 1,
				Delay:           generics.GetFakeObj[time.Duration]().String(),
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

func (s *releasesTestSuite) TestGetComponentRelease() {
	release, err := s.apiClient.CreateComponentRelease(s.ctx, s.compID, &models.ServiceCreateComponentReleaseRequest{
		BuildID: s.buildID,
		Strategy: &models.ServiceCreateComponentReleaseRequestStrategy{
			InstallsPerStep: 1,
			Delay:           generics.GetFakeObj[time.Duration]().String(),
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

func (s *releasesTestSuite) TestGetComponentReleaseSteps() {
	release, err := s.apiClient.CreateComponentRelease(s.ctx, s.compID, &models.ServiceCreateComponentReleaseRequest{
		BuildID: s.buildID,
		Strategy: &models.ServiceCreateComponentReleaseRequestStrategy{
			InstallsPerStep: 1,
			Delay:           generics.GetFakeObj[time.Duration]().String(),
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

	s.T().Run("successful when installs per step is 0", func(t *testing.T) {
		release, err := s.apiClient.CreateComponentRelease(s.ctx, s.compID, &models.ServiceCreateComponentReleaseRequest{
			BuildID: s.buildID,
			Strategy: &models.ServiceCreateComponentReleaseRequestStrategy{
				InstallsPerStep: 1,
				Delay:           generics.GetFakeObj[time.Duration]().String(),
			},
		})
		require.NoError(t, err)
		require.NotEmpty(t, release)

		fetched, err := s.apiClient.GetReleaseSteps(s.ctx, release.ID)
		require.NoError(t, err)
		require.NotEmpty(t, fetched)

		require.Equal(t, fetched[0].RequestedInstallIds[0], s.installID)
	})
}
