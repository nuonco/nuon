package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// ---------------------------------------------------------------------------
// Success cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestGetAppComponentsSuccess() {
	testCases := []struct {
		name          string
		setupFunc     func() []string // Returns component IDs that were created
		queryParams   string
		expectedCount int
		validateFunc  func([]app.Component)
	}{
		{
			name: "returns seeded components when component ids in app config",
			setupFunc: func() []string {
				// testAppConfig already has 6 typed components from the full seed
				return s.testAppConfig.ComponentIDs
			},
			queryParams:   "",
			expectedCount: 6,
			validateFunc: func(components []app.Component) {
				assert.Len(s.T(), components, 6)
				for _, comp := range components {
					assert.Equal(s.T(), s.testApp.ID, comp.AppID)
				}
			},
		},
		{
			name: "filters by search query",
			setupFunc: func() []string {
				// Rename a seeded component so we can search for it
				err := s.deps.DB.WithContext(s.ctx).
					Model(&app.Component{ID: s.testAppConfig.ComponentIDs[0]}).
					Update("name", "searchable_component").Error
				require.NoError(s.T(), err)
				return s.testAppConfig.ComponentIDs
			},
			queryParams:   "?q=searchable",
			expectedCount: 1,
			validateFunc: func(components []app.Component) {
				assert.Len(s.T(), components, 1)
				assert.Contains(s.T(), components[0].Name, "searchable")
			},
		},
		{
			name: "filters by component type",
			setupFunc: func() []string {
				// testAppConfig has one terraform_module component (index 1) with a real config connection
				return s.testAppConfig.ComponentIDs
			},
			queryParams:   "?types=terraform_module",
			expectedCount: 1,
			validateFunc: func(components []app.Component) {
				assert.Len(s.T(), components, 1)
				assert.Equal(s.T(), app.ComponentTypeTerraformModule, components[0].Type)
			},
		},
		{
			name: "filters by component ids",
			setupFunc: func() []string {
				// Use the pre-seeded components; filter will request only the first two
				return s.testAppConfig.ComponentIDs
			},
			queryParams:   "", // Will be set dynamically in test
			expectedCount: 2,
			validateFunc: func(components []app.Component) {
				assert.Len(s.T(), components, 2)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			componentIDs := tc.setupFunc()

			// Build query params dynamically for component_ids filter test
			queryParams := tc.queryParams
			if tc.name == "filters by component ids" && len(componentIDs) >= 2 {
				queryParams = fmt.Sprintf("?component_ids=%s,%s", componentIDs[0], componentIDs[1])
			}

			path := fmt.Sprintf("/v1/apps/%s/components%s", s.testApp.ID, queryParams)
			rr := s.makeRequest(http.MethodGet, path, nil)

			if rr.Code != http.StatusOK {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusOK, rr.Code)

			var response []app.Component
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			require.NoError(s.T(), err)

			assert.Len(s.T(), response, tc.expectedCount)

			if tc.validateFunc != nil {
				tc.validateFunc(response)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Empty app case
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestGetAppComponentsEmptyApp() {
	s.Run("returns empty for app with no component ids in config", func() {
		// Use a fresh app with a bare config (no components) to test the empty case
		emptyApp := s.deps.Seeder.CreateApp(s.ctx, s.T())
		s.deps.Seeder.CreateBareAppConfig(s.ctx, s.T(), emptyApp.ID)

		path := fmt.Sprintf("/v1/apps/%s/components", emptyApp.ID)
		rr := s.makeRequest(http.MethodGet, path, nil)

		if rr.Code != http.StatusOK {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response []app.Component
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)

		assert.Len(s.T(), response, 0, "should return empty when component not in app config")
	})
}
