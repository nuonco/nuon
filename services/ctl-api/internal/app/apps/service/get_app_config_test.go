package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lib/pq"
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

func (s *AppConfigsTestSuite) TestGetAppConfigRecurseExcludesStoredForeignComponentIDs() {
	ctx := cctx.SetOrgIDContext(cctx.SetAccountContext(context.Background(), s.testAcc), s.testOrg.ID)
	otherCtx, _ := s.service.Seeder.EnsureAccount(context.Background(), s.T())
	otherCtx, _ = s.service.Seeder.EnsureOrg(otherCtx, s.T())
	otherOrgApp := s.service.Seeder.CreateApp(otherCtx, s.T())
	otherOrgCfg := s.service.Seeder.CreateBareAppConfig(otherCtx, s.T(), otherOrgApp.ID)
	foreign := s.service.Seeder.CreateComponent(otherCtx, s.T(), otherOrgApp.ID, app.ComponentTypeHelmChart)
	s.service.Seeder.CreateHelmComponentConfigConnection(otherCtx, s.T(), foreign.ID, otherOrgCfg.ID)

	s.service.Seeder.CreateBareAppConfig(ctx, s.T(), s.testApp.ID)
	cfg := s.service.Seeder.CreateBareAppConfig(ctx, s.T(), s.testApp.ID)
	require.NoError(s.T(), s.service.DB.Model(&app.AppConfig{}).Where("id = ?", cfg.ID).
		Update("component_ids", pq.StringArray{foreign.ID}).Error)

	rr := s.makeGetRequest(http.MethodGet, fmt.Sprintf("/v1/apps/%s/configs/%s?recurse=true", s.testApp.ID, cfg.ID))
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())

	var response app.AppConfig
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &response))
	assert.Empty(s.T(), response.ComponentConfigConnections)
}
