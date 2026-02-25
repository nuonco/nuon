package service

import (
	"encoding/json"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *GeneralPublicTestSuite) TestGetCloudPlatformRegions() {
	testCases := []struct {
		name             string
		cloudPlatform    string
		expectedCode     int
		validateResponse func(regions []app.CloudPlatformRegion)
	}{
		{
			name:          "returns AWS regions",
			cloudPlatform: "aws",
			expectedCode:  http.StatusOK,
			validateResponse: func(regions []app.CloudPlatformRegion) {
				assert.NotEmpty(s.T(), regions)
				// Verify at least one region has expected fields
				if len(regions) > 0 {
					assert.NotEmpty(s.T(), regions[0].Name)
					assert.NotEmpty(s.T(), regions[0].Value)
				}
			},
		},
		{
			name:          "returns Azure regions",
			cloudPlatform: "azure",
			expectedCode:  http.StatusOK,
			validateResponse: func(regions []app.CloudPlatformRegion) {
				assert.NotEmpty(s.T(), regions)
				if len(regions) > 0 {
					assert.NotEmpty(s.T(), regions[0].Name)
					assert.NotEmpty(s.T(), regions[0].Value)
				}
			},
		},
		{
			name:          "returns error for unsupported GCP platform",
			cloudPlatform: "gcp",
			expectedCode:  http.StatusBadRequest,
			validateResponse: func(regions []app.CloudPlatformRegion) {
			},
		},
		{
			name:          "returns error for invalid platform",
			cloudPlatform: "invalid-platform",
			expectedCode:  http.StatusBadRequest,
			validateResponse: func(regions []app.CloudPlatformRegion) {
				// Error response - no regions to validate
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			path := "/v1/general/cloud-platform/" + tc.cloudPlatform + "/regions"
			rr := s.makeRequest(http.MethodGet, path, nil)

			if rr.Code != tc.expectedCode {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedCode, rr.Code)

			if tc.expectedCode == http.StatusOK {
				var regions []app.CloudPlatformRegion
				err := json.Unmarshal(rr.Body.Bytes(), &regions)
				require.NoError(s.T(), err)
				tc.validateResponse(regions)
			}
		})
	}
}
