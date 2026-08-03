package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/appconfigsync"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

// ensureAppQueue gives the test app the queue the sync signal is enqueued onto.
func (s *AppConfigsTestSuite) ensureAppQueue() *app.Queue {
	q := &app.Queue{
		OrgID:       &s.testOrg.ID,
		OwnerID:     s.testApp.ID,
		OwnerType:   "apps",
		MaxDepth:    100,
		MaxInFlight: 1,
		CreatedByID: s.testAcc.ID,
	}
	require.NoError(s.T(), s.service.DB.Create(q).Error)
	return q
}

func (s *AppConfigsTestSuite) createConfigWithIntermediate() *app.AppConfig {
	ctx := cctx.SetAccountContext(context.Background(), s.testAcc)
	ctx = cctx.SetOrgIDContext(ctx, s.testOrg.ID)

	intermediate, err := json.Marshal(config.AppConfig{Version: "1"})
	require.NoError(s.T(), err)

	appConfig, err := s.service.AppsService.createAppConfig(ctx, s.testOrg.ID, s.testApp.ID, &CreateAppConfigRequest{
		IntermediateConfigJSON: string(intermediate),
	})
	require.NoError(s.T(), err)
	return appConfig
}

func (s *AppConfigsTestSuite) TestSyncAppConfigEnqueuesSyncSignal() {
	s.ensureAppQueue()
	appConfig := s.createConfigWithIntermediate()
	tests.ClearQueueSignals(s.T(), s.service.DB)

	path := fmt.Sprintf("/v1/apps/%s/configs/%s/sync", s.testApp.ID, appConfig.ID)
	rr := s.makeRequestWithBody(http.MethodPost, path, nil)

	if rr.Code != http.StatusAccepted {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusAccepted, rr.Code)

	signals := tests.GetQueueSignals(s.T(), s.service.DB)
	require.Len(s.T(), signals, 1, "expected one sync signal")
	assert.Equal(s.T(), appconfigsync.SignalType, signals[0].Type)

	sig, ok := signals[0].Signal.Signal.(*appconfigsync.Signal)
	require.True(s.T(), ok, "signal should be an app-config-sync signal")
	assert.Equal(s.T(), s.testApp.ID, sig.AppID)
	assert.Equal(s.T(), appConfig.ID, sig.AppConfigID)
}

// A config created without an intermediate config has nothing to apply, so the
// request must be rejected rather than enqueueing a sync that cannot succeed.
func (s *AppConfigsTestSuite) TestSyncAppConfigRejectsConfigWithoutIntermediateConfig() {
	s.ensureAppQueue()

	ctx := cctx.SetAccountContext(context.Background(), s.testAcc)
	ctx = cctx.SetOrgIDContext(ctx, s.testOrg.ID)
	appConfig, err := s.service.AppsService.createAppConfig(ctx, s.testOrg.ID, s.testApp.ID, &CreateAppConfigRequest{})
	require.NoError(s.T(), err)
	tests.ClearQueueSignals(s.T(), s.service.DB)

	path := fmt.Sprintf("/v1/apps/%s/configs/%s/sync", s.testApp.ID, appConfig.ID)
	rr := s.makeRequestWithBody(http.MethodPost, path, nil)

	assert.NotEqual(s.T(), http.StatusAccepted, rr.Code)
	assert.Empty(s.T(), tests.GetQueueSignals(s.T(), s.service.DB), "no sync signal should be enqueued")
}

// Two concurrent syncs of the same config would race on the same records.
func (s *AppConfigsTestSuite) TestSyncAppConfigRejectsConfigAlreadySyncing() {
	s.ensureAppQueue()
	appConfig := s.createConfigWithIntermediate()
	require.NoError(s.T(), s.service.DB.
		Model(&app.AppConfig{}).
		Where("id = ?", appConfig.ID).
		Update("status", app.AppConfigStatusSyncing).Error)
	tests.ClearQueueSignals(s.T(), s.service.DB)

	path := fmt.Sprintf("/v1/apps/%s/configs/%s/sync", s.testApp.ID, appConfig.ID)
	rr := s.makeRequestWithBody(http.MethodPost, path, nil)

	assert.NotEqual(s.T(), http.StatusAccepted, rr.Code)
	assert.Empty(s.T(), tests.GetQueueSignals(s.T(), s.service.DB), "no sync signal should be enqueued")
}

// A config belonging to another app must not be syncable through this app.
func (s *AppConfigsTestSuite) TestSyncAppConfigRejectsConfigFromAnotherApp() {
	s.ensureAppQueue()
	appConfig := s.createConfigWithIntermediate()

	otherApp := &app.App{
		Name:        "other-test-app",
		OrgID:       s.testOrg.ID,
		CreatedByID: s.testAcc.ID,
		Status:      app.AppStatusProvisioning,
	}
	require.NoError(s.T(), s.service.DB.Create(otherApp).Error)
	tests.ClearQueueSignals(s.T(), s.service.DB)

	path := fmt.Sprintf("/v1/apps/%s/configs/%s/sync", otherApp.ID, appConfig.ID)
	rr := s.makeRequestWithBody(http.MethodPost, path, nil)

	assert.NotEqual(s.T(), http.StatusAccepted, rr.Code)
	assert.Empty(s.T(), tests.GetQueueSignals(s.T(), s.service.DB), "no sync signal should be enqueued")
}
