package service

import (
	"encoding/json"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *GeneralPublicTestSuite) TestGetCLIConfig() {
	s.Run("returns CLI configuration", func() {
		rr := s.makeRequest(http.MethodGet, "/v1/general/cli-config", nil)

		if rr.Code != http.StatusOK {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusOK, rr.Code)

		var cliConfig CLIConfig
		err := json.Unmarshal(rr.Body.Bytes(), &cliConfig)
		require.NoError(s.T(), err)

		// Verify response has expected fields (values will be from test config)
		assert.NotEmpty(s.T(), cliConfig.DashboardURL)
		assert.NotEmpty(s.T(), cliConfig.RootDomain)
		// NuonAuthEnabled can be true or false based on config
		assert.IsType(s.T(), false, cliConfig.NuonAuthEnabled)
	})
}
