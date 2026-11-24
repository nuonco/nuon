package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (s *AppConfigsTestSuite) TestUpdateAppConfigInstallsRejectsConfigOutsideApp() {
	ctx := cctx.SetOrgIDContext(cctx.SetAccountContext(context.Background(), s.testAcc), s.testOrg.ID)
	current := s.service.Seeder.CreateAppConfig(ctx, s.T(), s.testApp.ID)
	install := s.service.Seeder.CreateInstall(ctx, s.T(), s.testApp)
	sameOrgApp := s.service.Seeder.CreateApp(ctx, s.T())
	otherCtx, _ := s.service.Seeder.EnsureAccount(context.Background(), s.T())
	otherCtx, _ = s.service.Seeder.EnsureOrg(otherCtx, s.T())
	otherOrgApp := s.service.Seeder.CreateApp(otherCtx, s.T())

	storedConfigID := func() string {
		var id string
		require.NoError(s.T(), s.service.DB.Raw("SELECT app_config_id FROM installs WHERE id = ?", install.ID).Scan(&id).Error)
		return id
	}
	body := UpdateAppConfigInstallsRequest{UpdateAll: true}

	rejected := map[string]string{
		"other org":          s.service.Seeder.CreateBareAppConfig(otherCtx, s.T(), otherOrgApp.ID).ID,
		"same org other app": s.service.Seeder.CreateBareAppConfig(ctx, s.T(), sameOrgApp.ID).ID,
	}
	for name, cfgID := range rejected {
		for _, path := range []string{
			fmt.Sprintf("/v1/apps/%s/configs/%s/update-installs", s.testApp.ID, cfgID),
			fmt.Sprintf("/v1/apps/%s/config/%s/update-installs", s.testApp.ID, cfgID),
		} {
			s.Run(name+" "+path, func() {
				rr := s.makeRequestWithBody(http.MethodPost, path, body)
				require.Equal(s.T(), http.StatusNotFound, rr.Code, rr.Body.String())
				assert.Equal(s.T(), current.ID, storedConfigID())
			})
		}
	}

	next := s.service.Seeder.CreateBareAppConfig(ctx, s.T(), s.testApp.ID)
	rr := s.makeRequestWithBody(http.MethodPost, fmt.Sprintf("/v1/apps/%s/configs/%s/update-installs", s.testApp.ID, next.ID), body)
	require.Equal(s.T(), http.StatusOK, rr.Code, rr.Body.String())
	assert.Equal(s.T(), next.ID, storedConfigID())
}
