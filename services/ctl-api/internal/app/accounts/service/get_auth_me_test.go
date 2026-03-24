package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

// ---------------------------------------------------------------------------
// Success: basic auth/me with no identities
// ---------------------------------------------------------------------------

func (s *AccountsServiceTestSuite) TestGetAuthMeSuccess() {
	rr := s.makeRequest(http.MethodGet, "/v1/auth/me", nil)

	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var response AuthMeResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), s.testAcc.ID, response.Account.ID)
	assert.Equal(s.T(), s.testAcc.Email, response.Account.Email)
	assert.Equal(s.T(), s.testAcc.Subject, response.Account.Subject)
	assert.Equal(s.T(), s.testAcc.AccountType, response.Account.AccountType)
	assert.NotNil(s.T(), response.Identities)
	assert.Empty(s.T(), response.Identities)

	// Verify computed fields are present in the JSON response
	var raw map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &raw)
	require.NoError(s.T(), err)

	_, hasOrgIDs := raw["org_ids"]
	assert.True(s.T(), hasOrgIDs, "response should include org_ids field")

	_, hasPermissions := raw["permissions"]
	assert.True(s.T(), hasPermissions, "response should include permissions field")
}

// ---------------------------------------------------------------------------
// Success: auth/me with a single identity
// ---------------------------------------------------------------------------

func (s *AccountsServiceTestSuite) TestGetAuthMeWithSingleIdentity() {
	identity := app.AccountIdentity{
		AccountID:    s.testAcc.ID,
		ProviderType: app.ProviderTypeGoogle,
		Sub:          fmt.Sprintf("google-oauth2|%s", s.testAcc.ID),
		Name:         "Test User",
		Picture:      "https://example.com/photo.jpg",
	}
	err := s.service.DB.WithContext(s.ctx).Create(&identity).Error
	require.NoError(s.T(), err)

	rr := s.makeRequest(http.MethodGet, "/v1/auth/me", nil)

	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var response AuthMeResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), s.testAcc.ID, response.Account.ID)
	require.Len(s.T(), response.Identities, 1)
	assert.Equal(s.T(), app.ProviderTypeGoogle, response.Identities[0].ProviderType)
	assert.Equal(s.T(), "Test User", response.Identities[0].Name)
	assert.Equal(s.T(), "https://example.com/photo.jpg", response.Identities[0].Picture)
}

// ---------------------------------------------------------------------------
// Success: auth/me with multiple identities
// ---------------------------------------------------------------------------

func (s *AccountsServiceTestSuite) TestGetAuthMeWithMultipleIdentities() {
	identities := []app.AccountIdentity{
		{
			AccountID:    s.testAcc.ID,
			ProviderType: app.ProviderTypeGoogle,
			Sub:          fmt.Sprintf("google-oauth2|multi-%s", s.testAcc.ID),
			Name:         "Google User",
			Picture:      "https://google.com/photo.jpg",
		},
		{
			AccountID:    s.testAcc.ID,
			ProviderType: app.ProviderTypeGitHub,
			Sub:          fmt.Sprintf("github|multi-%s", s.testAcc.ID),
			Name:         "GitHub User",
			Picture:      "https://github.com/avatar.png",
		},
	}
	for i := range identities {
		err := s.service.DB.WithContext(s.ctx).Create(&identities[i]).Error
		require.NoError(s.T(), err)
	}

	rr := s.makeRequest(http.MethodGet, "/v1/auth/me", nil)

	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var response AuthMeResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	require.Len(s.T(), response.Identities, 2)

	// Collect provider types to verify both are present (order may vary)
	providerTypes := make(map[app.ProviderType]AuthMeIdentity)
	for _, ident := range response.Identities {
		providerTypes[ident.ProviderType] = ident
	}

	googleIdent, hasGoogle := providerTypes[app.ProviderTypeGoogle]
	require.True(s.T(), hasGoogle, "should include Google identity")
	assert.Equal(s.T(), "Google User", googleIdent.Name)
	assert.Equal(s.T(), "https://google.com/photo.jpg", googleIdent.Picture)

	githubIdent, hasGitHub := providerTypes[app.ProviderTypeGitHub]
	require.True(s.T(), hasGitHub, "should include GitHub identity")
	assert.Equal(s.T(), "GitHub User", githubIdent.Name)
	assert.Equal(s.T(), "https://github.com/avatar.png", githubIdent.Picture)
}

// ---------------------------------------------------------------------------
// Success: identity fields are filtered (only provider_type, name, picture)
// ---------------------------------------------------------------------------

func (s *AccountsServiceTestSuite) TestGetAuthMeIdentityFieldsFiltered() {
	identity := app.AccountIdentity{
		AccountID:    s.testAcc.ID,
		ProviderType: app.ProviderTypeOIDC,
		Sub:          fmt.Sprintf("oidc|%s", s.testAcc.ID),
		Name:         "OIDC User",
		Picture:      "https://idp.example.com/avatar.png",
	}
	err := s.service.DB.WithContext(s.ctx).Create(&identity).Error
	require.NoError(s.T(), err)

	rr := s.makeRequest(http.MethodGet, "/v1/auth/me", nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)

	// Parse as raw JSON to inspect the identity object keys
	var raw map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &raw)
	require.NoError(s.T(), err)

	identitiesRaw, ok := raw["identities"].([]interface{})
	require.True(s.T(), ok, "identities should be an array")
	require.Len(s.T(), identitiesRaw, 1)

	identityMap, ok := identitiesRaw[0].(map[string]interface{})
	require.True(s.T(), ok, "identity should be a JSON object")

	// Filtered fields should be present
	assert.Equal(s.T(), "oidc", identityMap["provider_type"])
	assert.Equal(s.T(), "OIDC User", identityMap["name"])
	assert.Equal(s.T(), "https://idp.example.com/avatar.png", identityMap["picture"])

	// Sensitive fields should NOT be present
	_, hasSub := identityMap["sub"]
	assert.False(s.T(), hasSub, "sub should not be exposed in auth/me response")

	_, hasAccountID := identityMap["account_id"]
	assert.False(s.T(), hasAccountID, "account_id should not be exposed in auth/me response")

	_, hasID := identityMap["id"]
	assert.False(s.T(), hasID, "id should not be exposed in auth/me response")

	_, hasIdentityProviderID := identityMap["identity_provider_id"]
	assert.False(s.T(), hasIdentityProviderID, "identity_provider_id should not be exposed in auth/me response")
}

// ---------------------------------------------------------------------------
// Success: auth/me does not return another account's identities
// ---------------------------------------------------------------------------

func (s *AccountsServiceTestSuite) TestGetAuthMeIsolatesIdentitiesByAccount() {
	// Create an identity for the test account
	ownIdentity := app.AccountIdentity{
		AccountID:    s.testAcc.ID,
		ProviderType: app.ProviderTypeGoogle,
		Sub:          fmt.Sprintf("google-oauth2|own-%s", s.testAcc.ID),
		Name:         "Own Account User",
		Picture:      "https://example.com/own.jpg",
	}
	err := s.service.DB.WithContext(s.ctx).Create(&ownIdentity).Error
	require.NoError(s.T(), err)

	// Create a different account and give it an identity
	otherAcc := testseed.BuildAccount()
	err = s.service.DB.WithContext(s.ctx).Create(otherAcc).Error
	require.NoError(s.T(), err)

	otherIdentity := app.AccountIdentity{
		AccountID:    otherAcc.ID,
		ProviderType: app.ProviderTypeGitHub,
		Sub:          fmt.Sprintf("github|other-%s", otherAcc.ID),
		Name:         "Other Account User",
		Picture:      "https://example.com/other.jpg",
	}
	err = s.service.DB.WithContext(s.ctx).Create(&otherIdentity).Error
	require.NoError(s.T(), err)

	rr := s.makeRequest(http.MethodGet, "/v1/auth/me", nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var response AuthMeResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	// Should only see own identity, not the other account's
	require.Len(s.T(), response.Identities, 1)
	assert.Equal(s.T(), "Own Account User", response.Identities[0].Name)
	assert.Equal(s.T(), app.ProviderTypeGoogle, response.Identities[0].ProviderType)
}
