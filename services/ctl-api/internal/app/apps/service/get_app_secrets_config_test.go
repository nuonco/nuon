package service

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// TestGetAppSecretsConfig tests the GetAppSecretsConfig endpoint.
func (s *AppConfigTypesTestSuite) TestGetAppSecretsConfig() {
	s.Run("returns not found when no config exists", func() {
		rr := s.makeRequest(http.MethodGet, "/v1/apps/"+s.testApp.ID+"/secrets-configs/nonexistent-id", nil)

		// The handler calls First() which returns record not found
		assert.Equal(s.T(), http.StatusNotFound, rr.Code)
	})

	s.Run("returns config after creation", func() {
		ctx := context.Background()
		ctx = cctx.SetAccountContext(ctx, s.testAcc)
		ctx = cctx.SetOrgContext(ctx, s.testOrg)

		cfg := &app.AppSecretsConfig{
			AppID:       s.testApp.ID,
			AppConfigID: s.testAppConfig.ID,
			OrgID:       s.testOrg.ID,
			Secrets: []app.AppSecretConfig{
				{
					AppID:       s.testApp.ID,
					AppConfigID: s.testAppConfig.ID,
					Name:        "test_secret",
					DisplayName: "Test Secret",
					Description: "Test description",
				},
			},
		}
		err := s.service.DB.WithContext(ctx).Create(cfg).Error
		require.NoError(s.T(), err)

		rr := s.makeRequest(http.MethodGet, "/v1/apps/"+s.testApp.ID+"/secrets-configs/"+cfg.ID, nil)

		if rr.Code != http.StatusOK {
			s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
		}
		require.Equal(s.T(), http.StatusOK, rr.Code)

		var response app.AppSecretsConfig
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(s.T(), err)
		assert.Equal(s.T(), cfg.ID, response.ID)
		assert.Equal(s.T(), s.testApp.ID, response.AppID)
	})
}
