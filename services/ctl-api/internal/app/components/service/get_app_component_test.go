package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *ComponentsServiceTestSuite) TestGetAppComponentSuccess() {
	seededComponentID := s.testAppConfig.ComponentConfigConnections[0].ComponentID

	var seededComp app.Component
	err := s.deps.DB.WithContext(s.ctx).First(&seededComp, "id = ?", seededComponentID).Error
	require.NoError(s.T(), err)

	testCases := []struct {
		name              string
		componentIDOrName string
		validateFunc      func(*app.Component)
	}{
		{
			name:              "returns component by id",
			componentIDOrName: seededComponentID,
			validateFunc: func(component *app.Component) {
				assert.Equal(s.T(), s.testApp.ID, component.AppID)
				assert.Equal(s.T(), seededComponentID, component.ID)
				assert.NotEmpty(s.T(), component.Name)
				assert.NotEmpty(s.T(), component.CreatedAt)
			},
		},
		{
			name:              "returns component by name",
			componentIDOrName: seededComp.Name,
			validateFunc: func(component *app.Component) {
				assert.Equal(s.T(), s.testApp.ID, component.AppID)
				assert.Equal(s.T(), seededComponentID, component.ID)
				assert.Equal(s.T(), seededComp.Name, component.Name)
			},
		},
		{
			name:              "includes correct fields in response",
			componentIDOrName: seededComponentID,
			validateFunc: func(component *app.Component) {
				assert.NotEmpty(s.T(), component.ID)
				assert.NotEmpty(s.T(), component.Name)
				assert.NotEmpty(s.T(), component.CreatedAt)
				assert.NotEmpty(s.T(), component.UpdatedAt)
				assert.Equal(s.T(), s.testApp.ID, component.AppID)
				assert.Equal(s.T(), s.testOrg.ID, component.OrgID)
				assert.NotEmpty(s.T(), component.Status)
				assert.NotEmpty(s.T(), component.Links)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			path := fmt.Sprintf("/v1/apps/%s/component/%s", s.testApp.ID, tc.componentIDOrName)
			rr := s.makeRequest(http.MethodGet, path, nil)

			if rr.Code != http.StatusOK {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusOK, rr.Code)

			var response app.Component
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			require.NoError(s.T(), err)

			if tc.validateFunc != nil {
				tc.validateFunc(&response)
			}
		})
	}
}

func (s *ComponentsServiceTestSuite) TestGetAppComponentNotFound() {
	testCases := []struct {
		name      string
		setupFunc func() (string, string)
	}{
		{
			name: "nonexistent id",
			setupFunc: func() (string, string) {
				return s.testApp.ID, "cmp_nonexistent00000000000"
			},
		},
		{
			name: "id from different app",
			setupFunc: func() (string, string) {
				app2 := s.deps.Seeder.CreateApp(s.ctx, s.T())

				comp2 := s.deps.Seeder.CreateComponent(s.ctx, s.T(), app2.ID, app.ComponentTypeTerraformModule)

				return s.testApp.ID, comp2.ID
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			appID, componentIDOrName := tc.setupFunc()

			path := fmt.Sprintf("/v1/apps/%s/component/%s", appID, componentIDOrName)
			rr := s.makeRequest(http.MethodGet, path, nil)

			if rr.Code != http.StatusNotFound {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusNotFound, rr.Code)
		})
	}
}
