package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *CreateAppActionConfigTestSuite) TestCreateActionConfigRejectsAppConfigFromAnotherApp() {
	ctx, _ := s.service.Seeder.EnsureAccount(context.Background(), s.T())
	otherOrgCtx, otherOrg := s.service.Seeder.EnsureOrg(ctx, s.T())
	otherOrgApp := s.service.Seeder.CreateApp(otherOrgCtx, s.T())
	otherOrgConfig := &app.AppConfig{
		ID:     domains.NewConfigID(),
		OrgID:  otherOrg.ID,
		AppID:  otherOrgApp.ID,
		Status: app.AppConfigStatusActive,
	}
	require.NoError(s.T(), s.service.DB.WithContext(otherOrgCtx).Create(otherOrgConfig).Error)

	sameOrgApp := s.service.Seeder.CreateApp(s.ctx, s.T())
	sameOrgConfig := s.createAppConfig(sameOrgApp.ID)

	action := s.createActionWorkflow(s.testApp.ID, "foreign-config-action")
	paths := []string{
		fmt.Sprintf("/v1/apps/%s/actions/%s/configs", s.testApp.ID, action.ID),
		fmt.Sprintf("/v1/action-workflows/%s/configs", action.ID),
	}
	foreignConfigs := map[string]string{
		"app config in another org":            otherOrgConfig.ID,
		"app config of another app in the org": sameOrgConfig.ID,
	}

	for _, path := range paths {
		for name, appConfigID := range foreignConfigs {
			s.Run(path+" rejects "+name, func() {
				rr := s.makeRequest(http.MethodPost, path, CreateActionWorkflowConfigRequest{
					AppConfigID: appConfigID,
					Triggers: []CreateActionWorkflowConfigTriggerRequest{
						{Type: app.ActionWorkflowTriggerTypeManual},
					},
					Steps: []CreateActionWorkflowConfigStepRequest{
						{Name: "step", InlineContents: "echo hi"},
					},
				})
				if rr.Code != http.StatusNotFound {
					s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
				}
				require.Equal(s.T(), http.StatusNotFound, rr.Code)
			})
		}
	}
}
