package service

import (
	"encoding/json"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *GeneralPublicTestSuite) TestGetCurrentUser() {
	s.Run("returns current user from context", func() {
		rr := s.makeRequest(http.MethodGet, "/v1/general/current-user", nil)

		if rr.Code != http.StatusOK {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusOK, rr.Code)

		var account app.Account
		err := json.Unmarshal(rr.Body.Bytes(), &account)
		require.NoError(s.T(), err)

		// Verify returned account matches test account
		assert.Equal(s.T(), s.testAcc.ID, account.ID)
		assert.Equal(s.T(), s.testAcc.Email, account.Email)
		assert.Equal(s.T(), s.testAcc.Subject, account.Subject)
	})
}
