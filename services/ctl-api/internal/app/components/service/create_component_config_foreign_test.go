package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *ComponentsServiceTestSuite) TestCreateComponentConfigRejectsAppConfigFromAnotherApp() {
	ctx, _ := s.deps.Seeder.EnsureAccount(context.Background(), s.T())
	otherOrgCtx, _ := s.deps.Seeder.EnsureOrg(ctx, s.T())
	otherOrgApp := s.deps.Seeder.CreateApp(otherOrgCtx, s.T())
	otherOrgConfig := s.deps.Seeder.CreateAppConfig(otherOrgCtx, s.T(), otherOrgApp.ID)

	sameOrgApp := s.deps.Seeder.CreateApp(s.ctx, s.T())
	sameOrgConfig := s.deps.Seeder.CreateAppConfig(s.ctx, s.T(), sameOrgApp.ID)

	comp := s.createComponent(s.testApp.ID, app.ComponentTypeJob)
	paths := []string{
		fmt.Sprintf("/v1/apps/%s/components/%s/configs/job", s.testApp.ID, comp.ID),
		fmt.Sprintf("/v1/components/%s/configs/job", comp.ID),
	}
	foreignConfigs := map[string]string{
		"app config in another org":            otherOrgConfig.ID,
		"app config of another app in the org": sameOrgConfig.ID,
	}

	for _, path := range paths {
		for name, appConfigID := range foreignConfigs {
			s.Run(path+" rejects "+name, func() {
				rr := s.makeRequest(http.MethodPost, path, CreateJobComponentConfigRequest{
					ImageURL:    "ubuntu",
					Tag:         "latest",
					AppConfigID: appConfigID,
				})
				if rr.Code != http.StatusNotFound {
					s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
				}
				require.Equal(s.T(), http.StatusNotFound, rr.Code)
			})
		}
	}
}
