package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// TestGetAppConfigV2Success tests GET /v1/apps/:app_id/configs/:config_id with existing config.
func (s *AppConfigsTestSuite) TestGetAppConfigV2Success() {
	ctx := context.Background()
	ctx = cctx.SetAccountContext(ctx, s.testAcc)
	ctx = cctx.SetOrgContext(ctx, s.testOrg)

	cfg := &app.AppConfig{
		ID:                domains.NewAppCfgID(),
		OrgID:             s.testOrg.ID,
		AppID:             s.testApp.ID,
		Status:            app.AppConfigStatusPending,
		StatusDescription: "pending",
		Readme:            "test readme",
		CLIVersion:        "1.0.0",
	}
	err := s.service.DB.WithContext(ctx).Create(cfg).Error
	require.NoError(s.T(), err)
	defer s.service.DB.Unscoped().Delete(&app.AppConfig{}, "id = ?", cfg.ID)

	path := fmt.Sprintf("/v1/apps/%s/configs/%s", s.testApp.ID, cfg.ID)
	rr := s.makeGetRequest(http.MethodGet, path)

	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var response models.AppAppConfig
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), cfg.ID, response.ID)
	assert.Equal(s.T(), s.testApp.ID, response.AppID)
	assert.Equal(s.T(), s.testOrg.ID, response.OrgID)
	assert.Equal(s.T(), "test readme", response.Readme)
	assert.Equal(s.T(), "1.0.0", response.CliVersion)
}

func (s *AppConfigsTestSuite) TestGetAppConfigV2NotFound() {
	nonExistentID := domains.NewAppCfgID()
	path := fmt.Sprintf("/v1/apps/%s/configs/%s", s.testApp.ID, nonExistentID)
	rr := s.makeGetRequest(http.MethodGet, path)

	if rr.Code != http.StatusNotFound {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
