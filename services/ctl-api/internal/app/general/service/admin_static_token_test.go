package service

import (
	"encoding/json"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *GeneralInternalTestSuite) TestAdminCreateStaticToken() {
	testCases := []struct {
		name           string
		requestBody    StaticTokenRequest
		expectedStatus int
		validateFunc   func(resp StaticTokenResponse)
	}{
		{
			name: "creates static token for existing account by email",
			requestBody: StaticTokenRequest{
				EmailOrSubject: s.testAcc.Email,
				Duration:       "8760h", // 1 year
			},
			expectedStatus: http.StatusCreated,
			validateFunc: func(resp StaticTokenResponse) {
				assert.NotEmpty(s.T(), resp.APIToken, "api_token should not be empty")
			},
		},
		{
			name: "creates static token for existing account by subject",
			requestBody: StaticTokenRequest{
				EmailOrSubject: s.testAcc.Subject,
				Duration:       "24h",
			},
			expectedStatus: http.StatusCreated,
			validateFunc: func(resp StaticTokenResponse) {
				assert.NotEmpty(s.T(), resp.APIToken, "api_token should not be empty")
			},
		},
		{
			name: "fails with missing email_or_subject",
			requestBody: StaticTokenRequest{
				Duration: "8760h",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "fails with missing duration",
			requestBody: StaticTokenRequest{
				EmailOrSubject: s.testAcc.Email,
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Make request
			rr := s.makeRequest(http.MethodPost, "/v1/general/admin-static-token", tc.requestBody)

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			if tc.expectedStatus == http.StatusCreated {
				// Unmarshal response
				var resp StaticTokenResponse
				err := json.Unmarshal(rr.Body.Bytes(), &resp)
				require.NoError(s.T(), err)

				// Validate response
				if tc.validateFunc != nil {
					tc.validateFunc(resp)
				}
			}
		})
	}
}
