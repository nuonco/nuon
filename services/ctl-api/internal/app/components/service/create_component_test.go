package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals"
)

// ---------------------------------------------------------------------------
// Success cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestCreateComponentSuccess() {
	testCases := []struct {
		name         string
		body         interface{}
		expectedName string
		validateFunc func(*app.Component)
	}{
		{
			name: "basic component with name only",
			body: CreateComponentRequest{
				Name: "my_component",
			},
			expectedName: "my_component",
			validateFunc: func(component *app.Component) {
				assert.Equal(s.T(), "my_component", component.Name)
				assert.Equal(s.T(), s.testApp.ID, component.AppID)
				assert.NotEmpty(s.T(), component.ID)
				assert.NotEmpty(s.T(), component.CreatedAt)
				assert.Equal(s.T(), app.ComponentStatus("queued"), component.Status)
				assert.Empty(s.T(), component.VarName)
			},
		},
		{
			name: "component with name and var_name",
			body: CreateComponentRequest{
				Name:    "var_name_component",
				VarName: "custom_var_name",
			},
			expectedName: "var_name_component",
			validateFunc: func(component *app.Component) {
				assert.Equal(s.T(), "var_name_component", component.Name)
				assert.Equal(s.T(), "custom_var_name", component.VarName)
				assert.Equal(s.T(), s.testApp.ID, component.AppID)
				assert.Equal(s.T(), app.ComponentStatus("queued"), component.Status)
			},
		},
		{
			name: "component with empty dependencies array",
			body: CreateComponentRequest{
				Name:         "empty_deps",
				Dependencies: []string{},
			},
			expectedName: "empty_deps",
			validateFunc: func(component *app.Component) {
				assert.Equal(s.T(), "empty_deps", component.Name)
				assert.Equal(s.T(), s.testApp.ID, component.AppID)
				assert.Equal(s.T(), app.ComponentStatus("queued"), component.Status)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			path := fmt.Sprintf("/v1/apps/%s/components", s.testApp.ID)
			rr := s.makeRequest(http.MethodPost, path, tc.body)

			if rr.Code != http.StatusCreated {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusCreated, rr.Code)

			var response app.Component
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			require.NoError(s.T(), err)

			assert.Equal(s.T(), tc.expectedName, response.Name)
			assert.Equal(s.T(), s.testApp.ID, response.AppID)

			// Verify persisted to database
			var dbComponent app.Component
			err = s.deps.DB.WithContext(s.ctx).First(&dbComponent, "id = ?", response.ID).Error
			require.NoError(s.T(), err)
			assert.Equal(s.T(), tc.expectedName, dbComponent.Name)
			assert.Equal(s.T(), app.ComponentStatus("queued"), dbComponent.Status)

			if tc.validateFunc != nil {
				tc.validateFunc(&response)
			}

		})
	}
}

// ---------------------------------------------------------------------------
// Validation error cases
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestCreateComponentValidationErrors() {
	testCases := []struct {
		name    string
		body    interface{}
		rawBody string
	}{
		{
			name: "nil body",
			body: nil,
		},
		{
			name: "empty JSON object",
			body: map[string]interface{}{},
		},
		{
			name: "missing name",
			body: map[string]interface{}{
				"var_name": "foo",
			},
		},
		{
			name: "empty name",
			body: map[string]interface{}{
				"name": "",
			},
		},
		{
			name: "name with spaces",
			body: map[string]interface{}{
				"name": "my component",
			},
		},
		{
			name: "name with uppercase",
			body: map[string]interface{}{
				"name": "MyComponent",
			},
		},
		{
			name: "name with hyphens",
			body: map[string]interface{}{
				"name": "my-component",
			},
		},
		{
			name:    "invalid JSON",
			rawBody: "{invalid json",
		},
		{
			name: "invalid var_name",
			body: map[string]interface{}{
				"name":     "valid_name",
				"var_name": "Bad Name",
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			path := fmt.Sprintf("/v1/apps/%s/components", s.testApp.ID)

			var rr *httptest.ResponseRecorder
			if tc.rawBody != "" {
				rr = s.makeRawRequest(http.MethodPost, path, tc.rawBody)
			} else {
				rr = s.makeRequest(http.MethodPost, path, tc.body)
			}

			if rr.Code != http.StatusBadRequest {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusBadRequest, rr.Code)
		})
	}
}

// ---------------------------------------------------------------------------
// Duplicate name
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestCreateComponentDuplicateName() {
	// Seed a component via seeder
	existingComponent := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeTerraformModule)

	// Try to create component with same name for same app
	reqBody := CreateComponentRequest{
		Name: existingComponent.Name,
	}

	path := fmt.Sprintf("/v1/apps/%s/components", s.testApp.ID)
	rr := s.makeRequest(http.MethodPost, path, reqBody)

	if rr.Code != http.StatusConflict {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusConflict, rr.Code)
}

// ---------------------------------------------------------------------------
// Component with dependencies
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestCreateComponentWithDependencies() {
	// Seed two components
	comp1 := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeTerraformModule)
	comp2 := s.deps.Seeder.CreateComponent(s.ctx, s.T(), s.testApp.ID, app.ComponentTypeTerraformModule)

	// Create a new component with dependencies
	reqBody := CreateComponentRequest{
		Name:         "new_comp",
		Dependencies: []string{comp1.Name, comp2.Name},
	}

	path := fmt.Sprintf("/v1/apps/%s/components", s.testApp.ID)
	rr := s.makeRequest(http.MethodPost, path, reqBody)

	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	var response app.Component
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	// Verify component_dependencies table has records
	var dependencies []app.ComponentDependency
	err = s.deps.DB.WithContext(s.ctx).
		Where("component_id = ?", response.ID).
		Find(&dependencies).Error
	require.NoError(s.T(), err)
	require.Len(s.T(), dependencies, 2, "expected 2 dependency records")

	// Verify the dependency IDs match
	depIDs := []string{dependencies[0].DependencyID, dependencies[1].DependencyID}
	assert.Contains(s.T(), depIDs, comp1.ID)
	assert.Contains(s.T(), depIDs, comp2.ID)

}

// ---------------------------------------------------------------------------
// Nonexistent dependency
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestCreateComponentNonexistentDependency() {
	reqBody := CreateComponentRequest{
		Name:         "foo",
		Dependencies: []string{"doesnt_exist"},
	}

	path := fmt.Sprintf("/v1/apps/%s/components", s.testApp.ID)
	rr := s.makeRequest(http.MethodPost, path, reqBody)

	require.Equal(s.T(), http.StatusBadRequest, rr.Code, "should not successfully create component with nonexistent dependency")
}

// ---------------------------------------------------------------------------
// Signals sent
// ---------------------------------------------------------------------------

func (s *ComponentsServiceTestSuite) TestCreateComponentSendsSignals() {
	// Reset mock
	s.mockEvClient.Reset()

	reqBody := CreateComponentRequest{
		Name: "signal_test_component",
	}

	path := fmt.Sprintf("/v1/apps/%s/components", s.testApp.ID)
	rr := s.makeRequest(http.MethodPost, path, reqBody)

	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	var response app.Component
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	// Verify 3 signals were sent
	capturedSignals := s.mockEvClient.GetSignals()
	require.Len(s.T(), capturedSignals, 3, "expected 3 signals")

	// All signals should target the created component
	for _, captured := range capturedSignals {
		assert.Equal(s.T(), response.ID, captured.ID, "signal should target the created component")
	}

	// Verify signal types
	var signalTypes []string
	for _, captured := range capturedSignals {
		sig, ok := captured.Signal.(*signals.Signal)
		require.True(s.T(), ok, "signal should be *signals.Signal")
		signalTypes = append(signalTypes, string(sig.Type))
	}

	assert.Contains(s.T(), signalTypes, string(signals.OperationCreated))
	assert.Contains(s.T(), signalTypes, string(signals.OperationProvision))
	assert.Contains(s.T(), signalTypes, string(signals.OperationPollDependencies))

}
