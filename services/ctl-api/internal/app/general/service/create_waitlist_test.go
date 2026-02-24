package service

import (
	"encoding/json"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *GeneralPublicTestSuite) TestCreateWaitlist() {
	s.Run("creates waitlist entry successfully", func() {
		req := WaitlistRequest{
			OrgName: "test-waitlist-org",
		}

		rr := s.makeRequest(http.MethodPost, "/v1/general/waitlist", req)

		if rr.Code != http.StatusCreated {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusCreated, rr.Code)

		var waitlist app.Waitlist
		err := json.Unmarshal(rr.Body.Bytes(), &waitlist)
		require.NoError(s.T(), err)

		// Verify response has expected fields
		assert.Equal(s.T(), "test-waitlist-org", waitlist.OrgName)
		assert.Equal(s.T(), s.testAcc.ID, waitlist.CreatedByID)
		assert.NotEmpty(s.T(), waitlist.ID)

		// Verify database state
		var dbWaitlist app.Waitlist
		err = s.service.DB.Where("id = ?", waitlist.ID).First(&dbWaitlist).Error
		require.NoError(s.T(), err)
		assert.Equal(s.T(), "test-waitlist-org", dbWaitlist.OrgName)
		assert.Equal(s.T(), s.testAcc.ID, dbWaitlist.CreatedByID)
	})

	s.Run("validation errors", func() {
		testCases := []struct {
			name        string
			request     interface{}
			description string
		}{
			{
				name:        "missing org_name",
				request:     map[string]string{},
				description: "org_name is required",
			},
			{
				name: "empty org_name",
				request: WaitlistRequest{
					OrgName: "",
				},
				description: "empty org_name should fail validation",
			},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				rr := s.makeRequest(http.MethodPost, "/v1/general/waitlist", tc.request)

				if rr.Code == http.StatusCreated {
					s.T().Logf("Expected error but got success. Status: %d, Body: %s", rr.Code, rr.Body.String())
				}

				// Should return an error status (400 or 500)
				assert.NotEqual(s.T(), http.StatusCreated, rr.Code, tc.description)
				assert.Contains(s.T(), []int{http.StatusBadRequest, http.StatusInternalServerError}, rr.Code)
			})
		}
	})
}
