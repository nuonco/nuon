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
// Success: no journeys on account
// ---------------------------------------------------------------------------

func (s *AccountsServiceTestSuite) TestGetUserJourneysEmpty() {
	rr := s.makeRequest(http.MethodGet, "/v1/account/user-journeys", nil)

	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var response app.UserJourneys
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	assert.Empty(s.T(), response)
}

// ---------------------------------------------------------------------------
// Success: single journey with partial completion
// ---------------------------------------------------------------------------

func (s *AccountsServiceTestSuite) TestGetUserJourneysSingleJourney() {
	journey := testseed.BuildUserJourney()
	journey.Steps[0].Complete = true
	s.testAcc.UserJourneys = app.UserJourneys{journey}
	err := s.service.DB.WithContext(s.ctx).Save(s.testAcc).Error
	require.NoError(s.T(), err)

	rr := s.makeRequest(http.MethodGet, "/v1/account/user-journeys", nil)

	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var response app.UserJourneys
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	require.Len(s.T(), response, 1)
	assert.Equal(s.T(), "onboarding", response[0].Name)
	assert.Equal(s.T(), "Getting Started", response[0].Title)
	require.Len(s.T(), response[0].Steps, 3)

	// First step should be complete, rest incomplete
	assert.True(s.T(), response[0].Steps[0].Complete)
	assert.False(s.T(), response[0].Steps[1].Complete)
	assert.False(s.T(), response[0].Steps[2].Complete)
}

// ---------------------------------------------------------------------------
// Success: multiple journeys
// ---------------------------------------------------------------------------

func (s *AccountsServiceTestSuite) TestGetUserJourneysMultipleJourneys() {
	onboarding := testseed.BuildCompletedUserJourney()
	advanced := testseed.BuildUserJourney()
	advanced.Name = "advanced-setup"
	advanced.Title = "Advanced Setup"
	advanced.Steps = []app.UserJourneyStep{
		{Name: "configure-runner", Title: "Configure Runner", Complete: false},
		{Name: "deploy-app", Title: "Deploy Application", Complete: false},
	}
	s.testAcc.UserJourneys = app.UserJourneys{onboarding, advanced}
	err := s.service.DB.WithContext(s.ctx).Save(s.testAcc).Error
	require.NoError(s.T(), err)

	rr := s.makeRequest(http.MethodGet, "/v1/account/user-journeys", nil)

	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var response app.UserJourneys
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	require.Len(s.T(), response, 2)

	// First journey: onboarding, fully complete
	assert.Equal(s.T(), "onboarding", response[0].Name)
	for _, step := range response[0].Steps {
		assert.True(s.T(), step.Complete, "onboarding step %q should be complete", step.Name)
	}

	// Second journey: advanced-setup, all incomplete
	assert.Equal(s.T(), "advanced-setup", response[1].Name)
	assert.Len(s.T(), response[1].Steps, 2)
	for _, step := range response[1].Steps {
		assert.False(s.T(), step.Complete, "advanced step %q should be incomplete", step.Name)
	}
}

// ---------------------------------------------------------------------------
// Journeys include step metadata
// ---------------------------------------------------------------------------

func (s *AccountsServiceTestSuite) TestGetUserJourneysIncludesMetadata() {
	journey := testseed.BuildUserJourney()
	journey.Steps[0].Complete = true
	journey.Steps[0].Metadata = map[string]interface{}{
		"org_id":   "org_12345",
		"org_name": "my-trial",
	}
	s.testAcc.UserJourneys = app.UserJourneys{journey}
	err := s.service.DB.WithContext(s.ctx).Save(s.testAcc).Error
	require.NoError(s.T(), err)

	rr := s.makeRequest(http.MethodGet, "/v1/account/user-journeys", nil)
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var response app.UserJourneys
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	require.Len(s.T(), response, 1)
	require.NotNil(s.T(), response[0].Steps[0].Metadata)
	assert.Equal(s.T(), "org_12345", response[0].Steps[0].Metadata["org_id"])
	assert.Equal(s.T(), "my-trial", response[0].Steps[0].Metadata["org_name"])
}
