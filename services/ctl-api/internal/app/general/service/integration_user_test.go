package service

import (
	"encoding/json"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *GeneralInternalTestSuite) TestCreateIntegrationUser() {
	testCases := []struct {
		name           string
		expectedStatus int
		validateFunc   func(resp CreateIntegrationUserResponse)
	}{
		{
			name:           "creates integration user with api token",
			expectedStatus: http.StatusCreated,
			validateFunc: func(resp CreateIntegrationUserResponse) {
				assert.NotEmpty(s.T(), resp.APIToken, "api_token should not be empty")
				assert.NotEmpty(s.T(), resp.Email, "email should not be empty")
				assert.Contains(s.T(), resp.Email, "@nuon.co", "email should be nuon.co domain")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Make request
			rr := s.makeRequest(http.MethodPost, "/v1/general/integration-user", map[string]interface{}{})

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			// Unmarshal response
			var resp CreateIntegrationUserResponse
			err := json.Unmarshal(rr.Body.Bytes(), &resp)
			require.NoError(s.T(), err)

			// Validate response
			if tc.validateFunc != nil {
				tc.validateFunc(resp)
			}
		})
	}
}
