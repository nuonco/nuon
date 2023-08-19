package integration

import (
	"os"
	"testing"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type sandboxesTestSuite struct {
	baseIntegrationTestSuite

	orgID string
	appID string
}

func TestSandboxesSuite(t *testing.T) {
	integration := os.Getenv("INTEGRATION")
	if integration == "" {
		t.Skip("INTEGRATION=true must be set in environment to run.")
	}

	suite.Run(t, new(sandboxesTestSuite))
}

// NOTE: sandboxes are seeded and managed using the internal api, so we just expect that they exist
func (s *sandboxesTestSuite) TestGetSandboxes() {
	s.T().Run("success", func(t *testing.T) {
		sandboxes, err := s.apiClient.GetSandboxes(s.ctx)
		require.NoError(t, err)
		require.NotEmpty(t, sandboxes)

		// make sure one of the sandboxes has `aws-eks` because that is what we seed
		for _, sandbox := range sandboxes {
			if sandbox.Name == "aws-eks" {
				return
			}
		}
		require.True(t, false, "no sandbox named aws-eks found")
	})
}

func (s *sandboxesTestSuite) TestGetSandbox() {
	sandboxes, err := s.apiClient.GetSandboxes(s.ctx)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), sandboxes)

	s.T().Run("success", func(t *testing.T) {
		sandbox, err := s.apiClient.GetSandbox(s.ctx, sandboxes[0].ID)
		require.NoError(t, err)
		require.Equal(t, sandboxes[0].ID, sandbox.ID)
	})

	s.T().Run("errors on invalid sandbox", func(t *testing.T) {
		sandbox, err := s.apiClient.GetSandbox(s.ctx, generics.GetFakeObj[string]())
		require.Error(t, err)
		require.Nil(t, sandbox)
	})
}

func (s *sandboxesTestSuite) TestGetSandboxReleases() {
	sandboxes, err := s.apiClient.GetSandboxes(s.ctx)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), sandboxes)

	s.T().Run("success", func(t *testing.T) {
		releases, err := s.apiClient.GetSandboxReleases(s.ctx, sandboxes[0].ID)
		require.NoError(t, err)
		require.NotEmpty(t, releases)
	})

	s.T().Run("errors on invalid sandbox", func(t *testing.T) {
		releases, err := s.apiClient.GetSandboxReleases(s.ctx, generics.GetFakeObj[string]())
		require.Error(t, err)
		require.Nil(t, releases)
	})
}
