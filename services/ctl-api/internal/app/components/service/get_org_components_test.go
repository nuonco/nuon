package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *ComponentsServiceTestSuite) TestGetOrgComponentsSuccess() {
	testCases := []struct {
		name          string
		queryParams   string
		expectedCount int
		validateFunc  func([]*models.AppComponent)
	}{
		{
			name:          "returns components for the org",
			queryParams:   "",
			expectedCount: 6,
			validateFunc: func(comps []*models.AppComponent) {
				s.Assert().GreaterOrEqual(len(comps), 6, "should have at least 6 components from full app config")
			},
		},
		{
			name:          "filters by component_ids",
			queryParams:   "", // set dynamically below
			expectedCount: 2,
			validateFunc: func(comps []*models.AppComponent) {
				s.Assert().Equal(2, len(comps), "should return exactly 2 filtered components")
			},
		},
		{
			name:          "returns empty for bogus component_ids filter",
			queryParams:   "?component_ids=cmpbogusid1,cmpbogusid2",
			expectedCount: 0,
			validateFunc: func(comps []*models.AppComponent) {
				s.Assert().Equal(0, len(comps), "should return empty array for non-existent component IDs")
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			queryParams := tc.queryParams
			if tc.name == "filters by component_ids" {
				queryParams = fmt.Sprintf("?component_ids=%s,%s",
					s.testAppConfig.ComponentIDs[0],
					s.testAppConfig.ComponentIDs[1])
			}

			rr := s.makeRequest(http.MethodGet, "/v1/components"+queryParams, nil)

			if rr.Code != http.StatusOK {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			s.Require().Equal(http.StatusOK, rr.Code)

			var result []*models.AppComponent
			err := json.NewDecoder(rr.Body).Decode(&result)
			s.Require().NoError(err)

			if tc.validateFunc != nil {
				tc.validateFunc(result)
			}
		})
	}
}
