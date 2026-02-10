package service

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// TestGetAppPoliciesConfigs tests the GetAppPoliciesConfigs endpoint.
func (s *AppConfigTypesTestSuite) TestGetAppPoliciesConfigs() {
	s.Run("returns empty array when no configs exist", func() {
		rr := s.makeRequest(http.MethodGet, "/v1/apps/"+s.testApp.ID+"/policies-configs", nil)

		if rr.Code != http.StatusOK {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response []app.AppPoliciesConfig
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)
		assert.Empty(s.T(), response)
	})

	s.Run("returns configs after creation", func() {
		ctx := context.Background()
		ctx = cctx.SetAccountContext(ctx, s.testAcc)
		ctx = cctx.SetOrgContext(ctx, s.testOrg)

		cfg := &app.AppPoliciesConfig{
			AppID:       s.testApp.ID,
			AppConfigID: s.testAppConfig.ID,
			OrgID:       s.testOrg.ID,
			Policies: []app.AppPolicyConfig{
				{
					AppID:       s.testApp.ID,
					AppConfigID: s.testAppConfig.ID,
					Type:        config.AppPolicyTypeKubernetesCluster,
					Contents:    "package test",
				},
			},
		}
		err := s.service.DB.WithContext(ctx).Create(cfg).Error
		require.NoError(s.T(), err)

		rr := s.makeRequest(http.MethodGet, "/v1/apps/"+s.testApp.ID+"/policies-configs", nil)

		if rr.Code != http.StatusOK {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response []app.AppPoliciesConfig
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)
		assert.Len(s.T(), response, 1)
		assert.Equal(s.T(), cfg.ID, response[0].ID)
	})
}
