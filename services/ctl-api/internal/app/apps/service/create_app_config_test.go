package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
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
	dbCtx := blobstore.WithBlobService(context.Background(), s.service.AppsService.blobSvc)
	err = s.service.DB.WithContext(dbCtx).First(&dbConfig, "id = ?", response.ID).Error
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

func (s *AppConfigsTestSuite) TestActivatingAPIAppConfigSyncsTriggers() {
	ctx := cctx.SetAccountContext(context.Background(), s.testAcc)
	ctx = cctx.SetOrgIDContext(ctx, s.testOrg.ID)
	trigger := &app.Trigger{OrgID: s.testOrg.ID, IngressKeyHash: "test-ingress-key-hash", Name: "pubsub", AuthType: app.TriggerAuthTypeNone, Envelope: app.EventEnvelopeTypeNone, Status: app.TriggerStatusActive}
	require.NoError(s.T(), s.service.DB.WithContext(ctx).Create(trigger).Error)
	branch := &app.AppBranch{OrgID: s.testOrg.ID, AppID: s.testApp.ID, Name: "main", ManagedBy: app.AppBranchManagedByConfig}
	require.NoError(s.T(), s.service.DB.WithContext(ctx).Create(branch).Error)

	intermediate, err := json.Marshal(config.AppConfig{
		Branch: &config.AppBranchConfig{Name: "main"},
		Triggers: &config.TriggersConfig{Rules: []*config.TriggerRuleConfig{{
			Name: "deploy", Trigger: trigger.Name, EventTypes: []string{"INSERT"},
			Target: &config.TriggerTargetConfig{Type: "app_branch_run", AppBranch: branch.Name},
		}}},
	})
	require.NoError(s.T(), err)
	appConfig, err := s.service.AppsService.createAppConfig(ctx, s.testOrg.ID, s.testApp.ID, &CreateAppConfigRequest{IntermediateConfigJSON: string(intermediate)})
	require.NoError(s.T(), err)

	_, err = s.service.AppsService.updateAppConfig(ctx, s.testOrg.ID, s.testApp.ID, appConfig.ID, &UpdateAppConfigRequest{Status: app.AppConfigStatusActive, StatusDescription: "synced"})
	require.NoError(s.T(), err)
	var rule app.TriggerRule
	require.NoError(s.T(), s.service.DB.Where(app.TriggerRule{AppConfigID: appConfig.ID, Name: "deploy"}).First(&rule).Error)
	assert.Equal(s.T(), trigger.ID, rule.TriggerID)
	assert.Equal(s.T(), branch.ID, rule.AppBranchID)
	assert.True(s.T(), rule.Enabled)
}

func (s *AppConfigsTestSuite) TestActivatingAPIAppConfigRollsBackAllTriggersWhenLaterRuleFails() {
	ctx := cctx.SetAccountContext(context.Background(), s.testAcc)
	ctx = cctx.SetOrgIDContext(ctx, s.testOrg.ID)
	trigger := &app.Trigger{OrgID: s.testOrg.ID, IngressKeyHash: "rollback-test-ingress-key-hash", Name: "rollback-trigger", AuthType: app.TriggerAuthTypeNone, Envelope: app.EventEnvelopeTypeNone, Status: app.TriggerStatusActive}
	require.NoError(s.T(), s.service.DB.WithContext(ctx).Create(trigger).Error)
	branch := &app.AppBranch{OrgID: s.testOrg.ID, AppID: s.testApp.ID, Name: "rollback-main", ManagedBy: app.AppBranchManagedByConfig}
	require.NoError(s.T(), s.service.DB.WithContext(ctx).Create(branch).Error)

	intermediate, err := json.Marshal(config.AppConfig{
		Branch: &config.AppBranchConfig{Name: branch.Name},
		Triggers: &config.TriggersConfig{Rules: []*config.TriggerRuleConfig{
			{Name: "first", Trigger: trigger.Name, EventTypes: []string{"push"}, Target: &config.TriggerTargetConfig{Type: "app_branch_run", AppBranch: branch.Name}},
			{Name: "second", Trigger: "missing-trigger", EventTypes: []string{"push"}, Target: &config.TriggerTargetConfig{Type: "app_branch_run", AppBranch: branch.Name}},
		}},
	})
	require.NoError(s.T(), err)
	appConfig, err := s.service.AppsService.createAppConfig(ctx, s.testOrg.ID, s.testApp.ID, &CreateAppConfigRequest{IntermediateConfigJSON: string(intermediate)})
	require.NoError(s.T(), err)

	_, err = s.service.AppsService.updateAppConfig(ctx, s.testOrg.ID, s.testApp.ID, appConfig.ID, &UpdateAppConfigRequest{Status: app.AppConfigStatusActive, StatusDescription: "synced"})
	require.ErrorContains(s.T(), err, "missing-trigger")

	var ruleCount int64
	require.NoError(s.T(), s.service.DB.Model(&app.TriggerRule{}).Where(app.TriggerRule{AppConfigID: appConfig.ID}).Count(&ruleCount).Error)
	assert.Zero(s.T(), ruleCount)

	var persisted app.AppConfig
	require.NoError(s.T(), s.service.DB.Where(app.AppConfig{ID: appConfig.ID}).First(&persisted).Error)
	assert.NotEqual(s.T(), app.AppConfigStatusActive, persisted.Status)
}
