package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

func (s *AppConfigsTestSuite) TestUpdateAppConfigRejectsComponentIDsOutsideApp() {
	ctx := cctx.SetOrgIDContext(cctx.SetAccountContext(context.Background(), s.testAcc), s.testOrg.ID)
	cfg := s.service.Seeder.CreateBareAppConfig(ctx, s.T(), s.testApp.ID)
	own := s.service.Seeder.CreateComponent(ctx, s.T(), s.testApp.ID, app.ComponentTypeHelmChart)
	sameOrgApp := s.service.Seeder.CreateApp(ctx, s.T())
	sameOrgComponent := s.service.Seeder.CreateComponent(ctx, s.T(), sameOrgApp.ID, app.ComponentTypeHelmChart)
	otherCtx, _ := s.service.Seeder.EnsureAccount(context.Background(), s.T())
	otherCtx, _ = s.service.Seeder.EnsureOrg(otherCtx, s.T())
	otherOrgApp := s.service.Seeder.CreateApp(otherCtx, s.T())
	otherOrgComponent := s.service.Seeder.CreateComponent(otherCtx, s.T(), otherOrgApp.ID, app.ComponentTypeHelmChart)

	rejected := map[string][]string{
		"other org":          {own.ID, otherOrgComponent.ID},
		"same org other app": {sameOrgComponent.ID},
		"unknown":            {"cmpdoesnotexist00000000000"},
	}
	for name, ids := range rejected {
		for _, path := range []string{
			fmt.Sprintf("/v1/apps/%s/configs/%s", s.testApp.ID, cfg.ID),
			fmt.Sprintf("/v1/apps/%s/config/%s", s.testApp.ID, cfg.ID),
		} {
			s.Run(name+" "+path, func() {
				rr := s.makeRequestWithBody(http.MethodPatch, path, UpdateAppConfigRequest{ComponentIDs: ids})
				require.Equal(s.T(), http.StatusBadRequest, rr.Code, rr.Body.String())

				var stored app.AppConfig
				require.NoError(s.T(), s.service.DB.First(&stored, "id = ?", cfg.ID).Error)
				assert.Empty(s.T(), stored.ComponentIDs)
			})
		}
	}

	rr := s.makeRequestWithBody(http.MethodPatch, fmt.Sprintf("/v1/apps/%s/configs/%s", s.testApp.ID, cfg.ID),
		UpdateAppConfigRequest{ComponentIDs: []string{own.ID, own.ID}})
	require.Equal(s.T(), http.StatusCreated, rr.Code, rr.Body.String())
	var stored app.AppConfig
	require.NoError(s.T(), s.service.DB.First(&stored, "id = ?", cfg.ID).Error)
	assert.Equal(s.T(), pq.StringArray{own.ID, own.ID}, stored.ComponentIDs)
}
