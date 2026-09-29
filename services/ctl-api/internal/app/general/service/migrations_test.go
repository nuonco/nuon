package service

import (
	"encoding/json"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/migrations"
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
				assert.NotNil(s.T(), resp, "response should not be nil")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			rr := s.makeRequest(http.MethodGet, "/v1/general/migrations", nil)

			if rr.Code != tc.expectedStatus {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), tc.expectedStatus, rr.Code)

			var resp []*migrations.MigrationModel
			err := json.Unmarshal(rr.Body.Bytes(), &resp)
			require.NoError(s.T(), err)

			if tc.validateFunc != nil {
				tc.validateFunc(resp)
			}
		})
	}
}
