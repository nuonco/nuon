package integration

import (
	"os"
	"testing"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type generalIntegrationTestSuite struct {
	baseIntegrationTestSuite
}

func TestGeneralSuite(t *testing.T) {
	t.Parallel()

	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	suite.Run(t, new(generalIntegrationTestSuite))
}

func (s *generalIntegrationTestSuite) TestGetCurrentUser() {
	s.T().Run("success", func(t *testing.T) {
		user, err := s.apiClient.GetCurrentUser(s.ctx)
		require.NoError(t, err)
		require.NotNil(t, user)
		require.NotEmpty(t, user.Subject)
	})
}

func (s *generalIntegrationTestSuite) TestGetCloudPlatforms() {
	s.T().Run("azure", func(t *testing.T) {
		regions, err := s.apiClient.GetCloudPlatformRegions(s.ctx, "azure")
		require.NoError(t, err)
		require.NotNil(t, regions)
		require.NotEmpty(t, regions)

		// verify eastus is in there
		found := false
		for _, region := range regions {
			if region.Value == "eastus" {
				found = true
				break
			}
		}
		require.True(t, found)
	})

	s.T().Run("aws", func(t *testing.T) {
		regions, err := s.apiClient.GetCloudPlatformRegions(s.ctx, "aws")
		require.NoError(t, err)
		require.NotNil(t, regions)
		require.NotEmpty(t, regions)

		// verify us-east-1 is in there
		found := false
		for _, region := range regions {
			if region.Value == "us-east-1" {
				found = true
				break
			}
		}
		require.True(t, found)
	})

	s.T().Run("invalid", func(t *testing.T) {
		regions, err := s.apiClient.GetCloudPlatformRegions(s.ctx, generics.GetFakeObj[string]())
		require.Error(t, err)
		require.Nil(t, regions)
	})
}
