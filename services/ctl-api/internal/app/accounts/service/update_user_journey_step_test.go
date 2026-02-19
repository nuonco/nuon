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

func (s *AccountsServiceTestSuite) TestUpdateUserJourneyStepSuccess() {
	testCases := []struct {
		name             string
		setupComplete    bool // initial completion state of create-org step
		reqBody          UpdateUserJourneyStepRequest
		expectedComplete bool
	}{
		{
			name:             "mark incomplete step as complete",
			setupComplete:    false,
			reqBody:          UpdateUserJourneyStepRequest{Complete: true},
			expectedComplete: true,
		},
		{
			name:             "mark complete step as incomplete",
			setupComplete:    true,
			reqBody:          UpdateUserJourneyStepRequest{Complete: false},
			expectedComplete: false,
		},
		{
			name:          "mark complete with metadata",
			setupComplete: false,
			reqBody: UpdateUserJourneyStepRequest{
				Complete: true,
				Metadata: map[string]interface{}{
					"org_id":   "org_12345",
					"org_name": "my-trial",
				},
			},
			expectedComplete: true,
		},
		{
			name:             "mark complete without metadata",
			setupComplete:    false,
			reqBody:          UpdateUserJourneyStepRequest{Complete: true},
			expectedComplete: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Seed journey with the desired initial state
			journey := testseed.BuildUserJourney()
			journey.Steps[0].Complete = tc.setupComplete
			s.testAcc.UserJourneys = app.UserJourneys{journey}
			err := s.service.DB.WithContext(s.ctx).Save(s.testAcc).Error
			require.NoError(s.T(), err)

			rr := s.makeRequest(http.MethodPatch, "/v1/account/user-journeys/onboarding/steps/create-org", tc.reqBody)

			if rr.Code != http.StatusOK {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusOK, rr.Code)

			var response app.Account
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			require.NoError(s.T(), err)

			require.Len(s.T(), response.UserJourneys, 1)
			require.Len(s.T(), response.UserJourneys[0].Steps, 3)
			assert.Equal(s.T(), tc.expectedComplete, response.UserJourneys[0].Steps[0].Complete)

			// Verify metadata if it was sent
			if tc.reqBody.Metadata != nil {
				for k, v := range tc.reqBody.Metadata {
					assert.Equal(s.T(), v, response.UserJourneys[0].Steps[0].Metadata[k])
				}
			}

			// Verify persisted to database
			var dbAccount app.Account
			err = s.service.DB.WithContext(s.ctx).First(&dbAccount, "id = ?", s.testAcc.ID).Error
			require.NoError(s.T(), err)

			require.Len(s.T(), dbAccount.UserJourneys, 1)
			assert.Equal(s.T(), tc.expectedComplete, dbAccount.UserJourneys[0].Steps[0].Complete)
		})
	}
}

// ---------------------------------------------------------------------------
// Metadata merging preserves existing metadata
// ---------------------------------------------------------------------------

func (s *AccountsServiceTestSuite) TestUpdateStepMetadataMergesWithExisting() {
	journey := testseed.BuildUserJourney()
	journey.Steps[1].Metadata = map[string]interface{}{
		"started_at": "2024-01-01",
		"source":     "dashboard",
	}
	s.testAcc.UserJourneys = app.UserJourneys{journey}
	err := s.service.DB.WithContext(s.ctx).Save(s.testAcc).Error
	require.NoError(s.T(), err)

	reqBody := UpdateUserJourneyStepRequest{
		Complete: true,
		Metadata: map[string]interface{}{
			"app_id": "app_99999",
			"source": "api",
		},
	}
	rr := s.makeRequest(http.MethodPatch, "/v1/account/user-journeys/onboarding/steps/create-app", reqBody)

	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var response app.Account
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	stepMeta := response.UserJourneys[0].Steps[1].Metadata
	require.NotNil(s.T(), stepMeta)

	// Existing key preserved
	assert.Equal(s.T(), "2024-01-01", stepMeta["started_at"])
	// New key added
	assert.Equal(s.T(), "app_99999", stepMeta["app_id"])
	// Overwritten key
	assert.Equal(s.T(), "api", stepMeta["source"])
}

// ---------------------------------------------------------------------------
// Update without metadata preserves existing metadata
// ---------------------------------------------------------------------------

func (s *AccountsServiceTestSuite) TestUpdateStepWithoutMetadataPreservesExisting() {
	journey := testseed.BuildUserJourney()
	journey.Steps[1].Metadata = map[string]interface{}{
		"app_id": "app_existing",
	}
	s.testAcc.UserJourneys = app.UserJourneys{journey}
	err := s.service.DB.WithContext(s.ctx).Save(s.testAcc).Error
	require.NoError(s.T(), err)

	reqBody := UpdateUserJourneyStepRequest{
		Complete: true,
	}
	rr := s.makeRequest(http.MethodPatch, "/v1/account/user-journeys/onboarding/steps/create-app", reqBody)
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var response app.Account
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	stepMeta := response.UserJourneys[0].Steps[1].Metadata
	require.NotNil(s.T(), stepMeta)
	assert.Equal(s.T(), "app_existing", stepMeta["app_id"])
}

// ---------------------------------------------------------------------------
// Does not affect other steps or journeys
// ---------------------------------------------------------------------------

func (s *AccountsServiceTestSuite) TestUpdateStepDoesNotAffectOtherSteps() {
	s.testAcc.UserJourneys = app.UserJourneys{testseed.BuildUserJourney()}
	err := s.service.DB.WithContext(s.ctx).Save(s.testAcc).Error
	require.NoError(s.T(), err)

	reqBody := UpdateUserJourneyStepRequest{Complete: true}
	rr := s.makeRequest(http.MethodPatch, "/v1/account/user-journeys/onboarding/steps/create-org", reqBody)
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var response app.Account
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	require.Len(s.T(), response.UserJourneys[0].Steps, 3)
	assert.True(s.T(), response.UserJourneys[0].Steps[0].Complete, "create-org should be complete")
	assert.False(s.T(), response.UserJourneys[0].Steps[1].Complete, "create-app should remain incomplete")
	assert.False(s.T(), response.UserJourneys[0].Steps[2].Complete, "create-install should remain incomplete")
}

func (s *AccountsServiceTestSuite) TestUpdateStepDoesNotAffectOtherJourneys() {
	onboarding := testseed.BuildUserJourney()
	advanced := testseed.BuildUserJourney()
	advanced.Name = "advanced"
	advanced.Title = "Advanced Setup"
	advanced.Steps = []app.UserJourneyStep{
		{Name: "configure-runner", Title: "Configure Runner", Complete: false},
	}
	s.testAcc.UserJourneys = app.UserJourneys{onboarding, advanced}
	err := s.service.DB.WithContext(s.ctx).Save(s.testAcc).Error
	require.NoError(s.T(), err)

	reqBody := UpdateUserJourneyStepRequest{Complete: true}
	rr := s.makeRequest(http.MethodPatch, "/v1/account/user-journeys/onboarding/steps/create-org", reqBody)
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var response app.Account
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	require.Len(s.T(), response.UserJourneys, 2)
	assert.True(s.T(), response.UserJourneys[0].Steps[0].Complete, "onboarding step should be updated")
	assert.False(s.T(), response.UserJourneys[1].Steps[0].Complete, "advanced step should be untouched")
}

// ---------------------------------------------------------------------------
// Not found cases
// ---------------------------------------------------------------------------

func (s *AccountsServiceTestSuite) TestUpdateStepNotFound() {
	testCases := []struct {
		name        string
		seedJourney bool
		path        string
	}{
		{
			name:        "nonexistent journey name",
			seedJourney: true,
			path:        "/v1/account/user-journeys/nonexistent/steps/create-org",
		},
		{
			name:        "nonexistent step name",
			seedJourney: true,
			path:        "/v1/account/user-journeys/onboarding/steps/nonexistent",
		},
		{
			name:        "no journeys on account",
			seedJourney: false,
			path:        "/v1/account/user-journeys/onboarding/steps/create-org",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.seedJourney {
				s.testAcc.UserJourneys = app.UserJourneys{testseed.BuildUserJourney()}
			} else {
				s.testAcc.UserJourneys = app.UserJourneys{}
			}
			err := s.service.DB.WithContext(s.ctx).Save(s.testAcc).Error
			require.NoError(s.T(), err)

			reqBody := UpdateUserJourneyStepRequest{Complete: true}
			rr := s.makeRequest(http.MethodPatch, tc.path, reqBody)

			if rr.Code != http.StatusNotFound {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusNotFound, rr.Code)
		})
	}
}

// ---------------------------------------------------------------------------
// Validation errors
// ---------------------------------------------------------------------------

func (s *AccountsServiceTestSuite) TestUpdateStepValidationErrors() {
	s.testAcc.UserJourneys = app.UserJourneys{testseed.BuildUserJourney()}
	err := s.service.DB.WithContext(s.ctx).Save(s.testAcc).Error
	require.NoError(s.T(), err)

	testCases := []struct {
		name    string
		rawBody string
	}{
		{
			name:    "invalid JSON",
			rawBody: "{invalid json",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			rr := s.makeRawRequest(http.MethodPatch, "/v1/account/user-journeys/onboarding/steps/create-org", tc.rawBody)

			if rr.Code != http.StatusBadRequest {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusBadRequest, rr.Code)
		})
	}
}
