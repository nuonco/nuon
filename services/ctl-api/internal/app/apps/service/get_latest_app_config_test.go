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

// TestGetAppLatestConfigSuccess tests GET /v1/apps/:app_id/latest-config.
func (s *AppConfigsTestSuite) TestGetAppLatestConfigSuccess() {
	testCases := []struct {
		name         string
		setupFunc    func() []string
		validateFunc func(*models.AppAppConfig)
	}{
		{
			name: "returns latest when multiple configs exist",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)
				ctx = cctx.SetOrgContext(ctx, s.testOrg)

				cfg1 := &app.AppConfig{
					ID:                domains.NewAppCfgID(),
					OrgID:             s.testOrg.ID,
					AppID:             s.testApp.ID,
					Status:            app.AppConfigStatusPending,
					StatusDescription: "pending",
					Readme:            "older config",
					CLIVersion:        "1.0.0",
				}
				err := s.service.DB.WithContext(ctx).Create(cfg1).Error
				require.NoError(s.T(), err)

				cfg2 := &app.AppConfig{
					ID:                domains.NewAppCfgID(),
					OrgID:             s.testOrg.ID,
					AppID:             s.testApp.ID,
					Status:            app.AppConfigStatusActive,
					StatusDescription: "success",
					Readme:            "latest config",
					CLIVersion:        "1.1.0",
				}
				err = s.service.DB.WithContext(ctx).Create(cfg2).Error
				require.NoError(s.T(), err)

				return []string{cfg1.ID, cfg2.ID}
			},
			validateFunc: func(cfg *models.AppAppConfig) {
				assert.Equal(s.T(), "latest config", cfg.Readme)
				assert.Equal(s.T(), "1.1.0", cfg.CliVersion)
			},
		},
		{
			name: "returns single config when only one exists",
			setupFunc: func() []string {
				ctx := context.Background()
				ctx = cctx.SetAccountContext(ctx, s.testAcc)
				ctx = cctx.SetOrgContext(ctx, s.testOrg)

				cfg := &app.AppConfig{
					ID:                domains.NewAppCfgID(),
					OrgID:             s.testOrg.ID,
					AppID:             s.testApp.ID,
					Status:            app.AppConfigStatusPending,
					StatusDescription: "pending",
					Readme:            "only config",
					CLIVersion:        "1.0.0",
				}
				err := s.service.DB.WithContext(ctx).Create(cfg).Error
				require.NoError(s.T(), err)
				return []string{cfg.ID}
			},
			validateFunc: func(cfg *models.AppAppConfig) {
				assert.Equal(s.T(), "only config", cfg.Readme)
				assert.Equal(s.T(), "1.0.0", cfg.CliVersion)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			configIDs := tc.setupFunc()

			s.T().Cleanup(func() {
				for _, cfgID := range configIDs {
					capturedID := cfgID
					s.service.DB.Unscoped().Delete(&app.AppConfig{}, "id = ?", capturedID)
				}
			})

			path := fmt.Sprintf("/v1/apps/%s/latest-config", s.testApp.ID)
			rr := s.makeGetRequest(http.MethodGet, path)

			if rr.Code != http.StatusOK {
				s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
			}
			require.Equal(s.T(), http.StatusOK, rr.Code)

			var response models.AppAppConfig
			err := json.Unmarshal(rr.Body.Bytes(), &response)
			if err != nil {
				s.T().Logf("Unmarshal error. Body: %s", rr.Body.String())
			}
			require.NoError(s.T(), err)

			tc.validateFunc(&response)
		})
	}
}

func (s *AppConfigsTestSuite) TestGetAppLatestConfigNotFound() {
	path := fmt.Sprintf("/v1/apps/%s/latest-config", s.testApp.ID)
	rr := s.makeGetRequest(http.MethodGet, path)

	if rr.Code != http.StatusNotFound {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}
