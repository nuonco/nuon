package service

import (
	"encoding/json"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *GeneralPublicTestSuite) TestGetConfigSchema() {
	testCases := []struct {
		name         string
		queryParams  string
		expectedCode int
	}{
		{
			name:         "returns schema for action type",
			queryParams:  "?type=action",
			expectedCode: http.StatusOK,
		},
		{
			name:         "returns schema for helm type",
			queryParams:  "?type=helm",
			expectedCode: http.StatusOK,
		},
		{
			name:         "returns schema for terraform type",
			queryParams:  "?type=terraform",
			expectedCode: http.StatusOK,
		},
		{
			name:         "returns schema for full type",
			queryParams:  "?type=full",
			expectedCode: http.StatusOK,
		},
		{
			name:         "returns error when type is missing",
			queryParams:  "",
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "returns error for unknown schema type",
			queryParams:  "?type=unknown-type",
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			path := "/v1/general/config-schema" + tc.queryParams
			rr := s.makeRequest(http.MethodGet, path, nil)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusOK {
				// Verify response is valid JSON - schema structure varies by type
				var result map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &result)
				require.NoError(s.T(), err)
				// Just verify it's a non-empty valid JSON object
				assert.NotEmpty(s.T(), result, "Schema response should be non-empty")
			}
		})
	}
}
