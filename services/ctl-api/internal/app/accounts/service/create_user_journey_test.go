package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

// ---------------------------------------------------------------------------
// Success cases
// ---------------------------------------------------------------------------

func (s *AccountsServiceTestSuite) TestCreateUserJourneySuccess() {
	testCases := []struct {
		name          string
		body          interface{}
		expectedName  string
		expectedSteps int
	}{
		{
			name: "with multiple steps",
			body: CreateUserJourneyRequest{
				Name:  "onboarding",
				Title: "Getting Started",
				Steps: []CreateUserJourneyStepReq{
					{Name: "create-org", Title: "Create Organization"},
					{Name: "create-app", Title: "Create App"},
					{Name: "create-install", Title: "Create Install"},
				},
			},
			expectedName:  "onboarding",
			expectedSteps: 3,
		},
		{
			name: "with single step",
			body: CreateUserJourneyRequest{
				Name:  "quick-start",
				Title: "Quick Start",
				Steps: []CreateUserJourneyStepReq{
					{Name: "setup", Title: "Initial Setup"},
				},
			},
			expectedName:  "quick-start",
			expectedSteps: 1,
		},
		{
			name: "with empty steps array",
			body: map[string]interface{}{
				"name":  "empty-steps-journey",
				"title": "Empty Steps",
				"steps": []map[string]interface{}{},
			},
			expectedName:  "empty-steps-journey",
			expectedSteps: 0,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			rr := s.makeRequest(http.MethodPost, "/v1/account/user-journeys", tc.body)

			if rr.Code != http.StatusCreated {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusCreated, rr.Code)

			var response app.Account
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			require.NoError(s.T(), err)

			// Find the created journey in the response
			require.NotEmpty(s.T(), response.UserJourneys)
			var found *app.UserJourney
			for i := range response.UserJourneys {
				if response.UserJourneys[i].Name == tc.expectedName {
					found = &response.UserJourneys[i]
					break
				}
			}
			require.NotNil(s.T(), found, "journey %q not found in response", tc.expectedName)

			assert.Equal(s.T(), tc.expectedName, found.Name)
			assert.Len(s.T(), found.Steps, tc.expectedSteps)

			// Verify all steps are initialized as incomplete
			for _, step := range found.Steps {
				assert.False(s.T(), step.Complete, "step %q should be incomplete on creation", step.Name)
			}

			// Verify persisted to database
			var dbAccount app.Account
			err = s.service.DB.WithContext(s.ctx).First(&dbAccount, "id = ?", s.testAcc.ID).Error
			require.NoError(s.T(), err)

			var dbFound *app.UserJourney
			for i := range dbAccount.UserJourneys {
				if dbAccount.UserJourneys[i].Name == tc.expectedName {
					dbFound = &dbAccount.UserJourneys[i]
					break
				}
			}
			require.NotNil(s.T(), dbFound, "journey %q not found in database", tc.expectedName)
			assert.Len(s.T(), dbFound.Steps, tc.expectedSteps)
			for _, step := range dbFound.Steps {
				assert.False(s.T(), step.Complete, "db step %q should be incomplete", step.Name)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Steps with complete=true in request should still initialize as false
// ---------------------------------------------------------------------------

func (s *AccountsServiceTestSuite) TestCreateUserJourneyStepsAlwaysStartIncomplete() {
	// Send a request with complete=true via raw map — the handler ignores it
	reqBody := map[string]interface{}{
		"name":  "sneaky-journey",
		"title": "Sneaky",
		"steps": []map[string]interface{}{
			{"name": "s1", "title": "Step 1", "complete": true},
			{"name": "s2", "title": "Step 2", "complete": true},
		},
	}

	rr := s.makeRequest(http.MethodPost, "/v1/account/user-journeys", reqBody)

	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	var response app.Account
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	require.Len(s.T(), response.UserJourneys, 1)
	for _, step := range response.UserJourneys[0].Steps {
		assert.False(s.T(), step.Complete, "step %q should be initialized as incomplete regardless of request", step.Name)
	}
}

// ---------------------------------------------------------------------------
// Validation error cases
// ---------------------------------------------------------------------------

func (s *AccountsServiceTestSuite) TestCreateUserJourneyValidationErrors() {
	testCases := []struct {
		name    string
		body    interface{}
		rawBody string
	}{
		{
			name: "no data sent (nil body)",
			body: nil,
		},
		{
			name: "empty JSON object",
			body: map[string]interface{}{},
		},
		{
			name: "missing name",
			body: map[string]interface{}{
				"title": "Some Title",
				"steps": []map[string]interface{}{
					{"name": "s1", "title": "Step 1"},
				},
			},
		},
		{
			name: "missing title",
			body: map[string]interface{}{
				"name": "my-journey",
				"steps": []map[string]interface{}{
					{"name": "s1", "title": "Step 1"},
				},
			},
		},
		{
			name: "missing steps field entirely",
			body: map[string]interface{}{
				"name":  "my-journey",
				"title": "My Journey",
			},
		},
		{
			name: "step missing name",
			body: map[string]interface{}{
				"name":  "my-journey",
				"title": "My Journey",
				"steps": []map[string]interface{}{
					{"title": "Step Without Name"},
				},
			},
		},
		{
			name: "step missing title",
			body: map[string]interface{}{
				"name":  "my-journey",
				"title": "My Journey",
				"steps": []map[string]interface{}{
					{"name": "step-no-title"},
				},
			},
		},
		{
			name: "step missing both name and title",
			body: map[string]interface{}{
				"name":  "my-journey",
				"title": "My Journey",
				"steps": []map[string]interface{}{
					{},
				},
			},
		},
		{
			name: "multiple steps with one invalid",
			body: map[string]interface{}{
				"name":  "my-journey",
				"title": "My Journey",
				"steps": []map[string]interface{}{
					{"name": "valid-step", "title": "Valid Step"},
					{"name": "", "title": "Missing Name Step"},
				},
			},
		},
		{
			name: "name is empty string",
			body: map[string]interface{}{
				"name":  "",
				"title": "Some Title",
				"steps": []map[string]interface{}{
					{"name": "s1", "title": "Step 1"},
				},
			},
		},
		{
			name: "title is empty string",
			body: map[string]interface{}{
				"name":  "my-journey",
				"title": "",
				"steps": []map[string]interface{}{
					{"name": "s1", "title": "Step 1"},
				},
			},
		},
		{
			name:    "invalid JSON",
			rawBody: "{invalid json",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			var rr *httptest.ResponseRecorder
			if tc.rawBody != "" {
				rr = s.makeRawRequest(http.MethodPost, "/v1/account/user-journeys", tc.rawBody)
			} else {
				rr = s.makeRequest(http.MethodPost, "/v1/account/user-journeys", tc.body)
			}

			if rr.Code != http.StatusBadRequest {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusBadRequest, rr.Code)
		})
	}
}

// ---------------------------------------------------------------------------
// Duplicate journey name
// ---------------------------------------------------------------------------

func (s *AccountsServiceTestSuite) TestCreateUserJourneyDuplicateNameFails() {
	// Seed the account with an existing journey
	s.testAcc.UserJourneys = app.UserJourneys{testseed.BuildUserJourney()}
	err := s.service.DB.WithContext(s.ctx).Save(s.testAcc).Error
	require.NoError(s.T(), err)

	// Attempt to create another journey with the same name
	reqBody := CreateUserJourneyRequest{
		Name:  "onboarding",
		Title: "Duplicate Journey",
		Steps: []CreateUserJourneyStepReq{
			{Name: "some-step", Title: "Some Step"},
		},
	}

	rr := s.makeRequest(http.MethodPost, "/v1/account/user-journeys", reqBody)

	if rr.Code != http.StatusConflict {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusConflict, rr.Code)

	// Verify error message mentions the duplicate
	var errResp map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(s.T(), err)
	assert.Contains(s.T(), errResp["error"], "already exists")
}

// ---------------------------------------------------------------------------
// Appends to existing journeys
// ---------------------------------------------------------------------------

func (s *AccountsServiceTestSuite) TestCreateUserJourneyAppendsToExisting() {
	// Seed the account with an existing journey
	s.testAcc.UserJourneys = app.UserJourneys{testseed.BuildCompletedUserJourney()}
	err := s.service.DB.WithContext(s.ctx).Save(s.testAcc).Error
	require.NoError(s.T(), err)

	// Create a second journey with a different name
	reqBody := CreateUserJourneyRequest{
		Name:  "advanced-setup",
		Title: "Advanced Setup",
		Steps: []CreateUserJourneyStepReq{
			{Name: "configure-runner", Title: "Configure Runner"},
			{Name: "deploy-app", Title: "Deploy Application"},
		},
	}

	rr := s.makeRequest(http.MethodPost, "/v1/account/user-journeys", reqBody)

	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	var response app.Account
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	// Should now have both journeys
	require.Len(s.T(), response.UserJourneys, 2)

	// Verify the original journey is preserved
	assert.Equal(s.T(), "onboarding", response.UserJourneys[0].Name)
	for _, step := range response.UserJourneys[0].Steps {
		assert.True(s.T(), step.Complete, "original journey step %q should remain complete", step.Name)
	}

	// Verify the new journey
	assert.Equal(s.T(), "advanced-setup", response.UserJourneys[1].Name)
	assert.Len(s.T(), response.UserJourneys[1].Steps, 2)
	for _, step := range response.UserJourneys[1].Steps {
		assert.False(s.T(), step.Complete, "new journey step %q should be incomplete", step.Name)
	}
}
