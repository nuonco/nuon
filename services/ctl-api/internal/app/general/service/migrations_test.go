package service

import (
	"encoding/json"
	"net/http"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *GeneralInternalTestSuite) TestGetMigrations() {
	testCases := []struct {
		name           string
		expectedStatus int
		validateFunc   func(resp []*migrations.MigrationModel)
	}{
		{
			name:           "returns migrations successfully",
			expectedStatus: http.StatusOK,
			validateFunc: func(resp []*migrations.MigrationModel) {
				// Response should be a valid array (even if empty)
				assert.NotNil(s.T(), resp, "response should not be nil")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Make request
			rr := s.makeRequest(http.MethodGet, "/v1/general/migrations", nil)

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			// Unmarshal response
			var resp []*migrations.MigrationModel
			err := json.Unmarshal(rr.Body.Bytes(), &resp)
			require.NoError(s.T(), err)

			// Validate response
			if tc.validateFunc != nil {
				tc.validateFunc(resp)
			}
		})
	}
}
