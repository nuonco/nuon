package integration

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.NotEmpty(t, user.Subject)
	})
}
