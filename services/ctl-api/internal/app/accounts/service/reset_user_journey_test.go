package service

import (
	"encoding/json"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/tests/testseed"
)

// ---------------------------------------------------------------------------
// Success cases
// ---------------------------------------------------------------------------

func (s *AccountsServiceTestSuite) TestResetUserJourneySuccess() {
	testCases := []struct {
		name      string
		seedSteps []app.UserJourneyStep
	}{
		{
			name: "all steps complete",
			seedSteps: []app.UserJourneyStep{
				{Name: "create-org", Title: "Create Organization", Complete: true},
				{Name: "create-app", Title: "Create App", Complete: true},
				{Name: "create-install", Title: "Create Install", Complete: true},
			},
		},
		{
			name: "mixed completion status",
			seedSteps: []app.UserJourneyStep{
				{Name: "create-org", Title: "Create Organization", Complete: true},
				{Name: "create-app", Title: "Create App", Complete: false},
				{Name: "create-install", Title: "Create Install", Complete: true},
			},
		},
		{
			name: "already all incomplete (no-op)",
			seedSteps: []app.UserJourneyStep{
				{Name: "create-org", Title: "Create Organization", Complete: false},
				{Name: "create-app", Title: "Create App", Complete: false},
				{Name: "create-install", Title: "Create Install", Complete: false},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.testAcc.UserJourneys = app.UserJourneys{
				{Name: "onboarding", Title: "Getting Started", Steps: tc.seedSteps},
			}
			err := s.service.DB.WithContext(s.ctx).Save(s.testAcc).Error
			require.NoError(s.T(), err)

			rr := s.makeRequest(http.MethodPost, "/v1/account/user-journeys/onboarding/reset", nil)

			if rr.Code != http.StatusOK {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusOK, rr.Code)

			var response app.Account
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			require.NoError(s.T(), err)

			require.Len(s.T(), response.UserJourneys, 1)
			require.Len(s.T(), response.UserJourneys[0].Steps, len(tc.seedSteps))

			for _, step := range response.UserJourneys[0].Steps {
				assert.False(s.T(), step.Complete, "step %q should be incomplete after reset", step.Name)
			}

			// Verify persisted to database
			var dbAccount app.Account
			err = s.service.DB.WithContext(s.ctx).First(&dbAccount, "id = ?", s.testAcc.ID).Error
			require.NoError(s.T(), err)

			require.Len(s.T(), dbAccount.UserJourneys, 1)
			for _, step := range dbAccount.UserJourneys[0].Steps {
				assert.False(s.T(), step.Complete, "db step %q should be incomplete after reset", step.Name)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Reset preserves step metadata
// ---------------------------------------------------------------------------

func (s *AccountsServiceTestSuite) TestResetJourneyPreservesStepMetadata() {
	journey := testseed.BuildCompletedUserJourney()
	journey.Steps[0].Metadata = map[string]interface{}{
		"org_id":   "org_12345",
		"org_name": "my-trial",
	}
	journey.Steps[1].Metadata = map[string]interface{}{
		"app_id": "app_99999",
	}
	s.testAcc.UserJourneys = app.UserJourneys{journey}
	err := s.service.DB.WithContext(s.ctx).Save(s.testAcc).Error
	require.NoError(s.T(), err)

	rr := s.makeRequest(http.MethodPost, "/v1/account/user-journeys/onboarding/reset", nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var response app.Account
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	require.Len(s.T(), response.UserJourneys, 1)

	// Steps should be incomplete but metadata preserved
	for _, step := range response.UserJourneys[0].Steps {
		assert.False(s.T(), step.Complete, "step %q should be incomplete", step.Name)
	}
	assert.Equal(s.T(), "org_12345", response.UserJourneys[0].Steps[0].Metadata["org_id"])
	assert.Equal(s.T(), "app_99999", response.UserJourneys[0].Steps[1].Metadata["app_id"])
}

// ---------------------------------------------------------------------------
// Reset does not affect other journeys
// ---------------------------------------------------------------------------

func (s *AccountsServiceTestSuite) TestResetJourneyDoesNotAffectOtherJourneys() {
	onboarding := testseed.BuildCompletedUserJourney()
	advanced := testseed.BuildCompletedUserJourney()
	advanced.Name = "advanced"
	advanced.Title = "Advanced Setup"
	advanced.Steps = []app.UserJourneyStep{
		{Name: "configure-runner", Title: "Configure Runner", Complete: true},
		{Name: "deploy-app", Title: "Deploy Application", Complete: true},
	}
	s.testAcc.UserJourneys = app.UserJourneys{onboarding, advanced}
	err := s.service.DB.WithContext(s.ctx).Save(s.testAcc).Error
	require.NoError(s.T(), err)

	// Reset only the onboarding journey
	rr := s.makeRequest(http.MethodPost, "/v1/account/user-journeys/onboarding/reset", nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var response app.Account
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	require.Len(s.T(), response.UserJourneys, 2)

	// Onboarding steps should all be reset
	for _, step := range response.UserJourneys[0].Steps {
		assert.False(s.T(), step.Complete, "onboarding step %q should be incomplete after reset", step.Name)
	}

	// Advanced steps should remain complete
	for _, step := range response.UserJourneys[1].Steps {
		assert.True(s.T(), step.Complete, "advanced step %q should remain complete", step.Name)
	}
}

// ---------------------------------------------------------------------------
// Not found cases
// ---------------------------------------------------------------------------

func (s *AccountsServiceTestSuite) TestResetJourneyNotFound() {
	testCases := []struct {
		name        string
		seedJourney bool
		path        string
	}{
		{
			name:        "nonexistent journey name",
			seedJourney: true,
			path:        "/v1/account/user-journeys/nonexistent/reset",
		},
		{
			name:        "no journeys on account",
			seedJourney: false,
			path:        "/v1/account/user-journeys/onboarding/reset",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.seedJourney {
				s.testAcc.UserJourneys = app.UserJourneys{testseed.BuildCompletedUserJourney()}
			} else {
				s.testAcc.UserJourneys = app.UserJourneys{}
			}
			err := s.service.DB.WithContext(s.ctx).Save(s.testAcc).Error
			require.NoError(s.T(), err)

			rr := s.makeRequest(http.MethodPost, tc.path, nil)

			if rr.Code != http.StatusNotFound {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusNotFound, rr.Code)
		})
	}
}
