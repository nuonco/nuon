package service

import (
	"encoding/json"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// ---------------------------------------------------------------------------
// Success: basic account retrieval
// ---------------------------------------------------------------------------

func (s *AccountsServiceTestSuite) TestGetCurrentAccountSuccess() {
	rr := s.makeRequest(http.MethodGet, "/v1/account", nil)

	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var response app.Account
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), s.testAcc.ID, response.ID)
	assert.Equal(s.T(), s.testAcc.Email, response.Email)
	assert.Equal(s.T(), s.testAcc.Subject, response.Subject)
	assert.Equal(s.T(), s.testAcc.AccountType, response.AccountType)
	assert.NotZero(s.T(), response.CreatedAt)
	assert.NotZero(s.T(), response.UpdatedAt)

	// Verify computed fields are present in the JSON response
	var raw map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &raw)
	require.NoError(s.T(), err)

	_, hasOrgIDs := raw["org_ids"]
	assert.True(s.T(), hasOrgIDs, "response should include org_ids field")

	_, hasPermissions := raw["permissions"]
	assert.True(s.T(), hasPermissions, "response should include permissions field")

	_, hasRoles := raw["roles"]
	assert.True(s.T(), hasRoles, "response should include roles field")
}

// ---------------------------------------------------------------------------
// Success: identities are NOT leaked in the account response
// ---------------------------------------------------------------------------

func (s *AccountsServiceTestSuite) TestGetCurrentAccountDoesNotLeakIdentities() {
	// Create an identity for the account
	identity := app.AccountIdentity{
		AccountID:    s.testAcc.ID,
		ProviderType: app.ProviderTypeGoogle,
		Sub:          "google-oauth2|123456789",
		Name:         "Test User",
		Picture:      "https://example.com/photo.jpg",
	}
	err := s.service.DB.WithContext(s.ctx).Create(&identity).Error
	require.NoError(s.T(), err)

	rr := s.makeRequest(http.MethodGet, "/v1/account", nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)

	// Parse as raw JSON to verify identities field is not present
	var raw map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &raw)
	require.NoError(s.T(), err)

	// Account model has `json:"-"` on Identities, so it should not appear
	_, hasIdentities := raw["identities"]
	assert.False(s.T(), hasIdentities, "GET /v1/account should not include identities")
}
