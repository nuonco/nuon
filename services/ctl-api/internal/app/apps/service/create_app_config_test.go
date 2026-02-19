package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// TestCreateAppConfigV2Success tests POST /v1/apps/:app_id/configs with valid input.
func (s *AppConfigsTestSuite) TestCreateAppConfigV2Success() {
	req := CreateAppConfigRequest{
		Readme:     "test readme",
		CLIVersion: "1.0.0",
	}

	path := fmt.Sprintf("/v1/apps/%s/configs", s.testApp.ID)
	rr := s.makeRequestWithBody(http.MethodPost, path, req)

	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	var response models.AppAppConfig
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	assert.NotEmpty(s.T(), response.ID)
	assert.Equal(s.T(), s.testApp.ID, response.AppID)
	assert.Equal(s.T(), s.testOrg.ID, response.OrgID)
	assert.Equal(s.T(), "test readme", response.Readme)
	assert.Equal(s.T(), "1.0.0", response.CliVersion)
	assert.Equal(s.T(), models.AppAppConfigStatus(app.AppConfigStatusPending), response.Status)

	var dbConfig app.AppConfig
	err = s.service.DB.First(&dbConfig, "id = ?", response.ID).Error
	require.NoError(s.T(), err)
	assert.Equal(s.T(), s.testApp.ID, dbConfig.AppID)
	assert.Equal(s.T(), s.testOrg.ID, dbConfig.OrgID)
	assert.Equal(s.T(), "test readme", dbConfig.Readme)
	assert.Equal(s.T(), "1.0.0", dbConfig.CLIVersion)
	assert.Equal(s.T(), app.AppConfigStatusPending, dbConfig.Status)
}

func (s *AppConfigsTestSuite) TestCreateAppConfigV2WithEmptyFields() {
	req := CreateAppConfigRequest{}

	path := fmt.Sprintf("/v1/apps/%s/configs", s.testApp.ID)
	rr := s.makeRequestWithBody(http.MethodPost, path, req)

	if rr.Code != http.StatusCreated {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusCreated, rr.Code)

	var response models.AppAppConfig
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	assert.NotEmpty(s.T(), response.ID)
	assert.Equal(s.T(), s.testApp.ID, response.AppID)
	assert.Equal(s.T(), "", response.Readme)
	assert.Equal(s.T(), "", response.CliVersion)
}
