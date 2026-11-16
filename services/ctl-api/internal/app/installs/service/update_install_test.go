package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/tests"
)

func (s *InstallsServiceTestSuite) TestUpdateInstallName() {
	install := s.createTestInstall()

	body := map[string]interface{}{
		"name": "updated-name",
	}

	path := fmt.Sprintf("/v1/installs/%s", install.ID)
	rr := s.makeRequest(http.MethodPatch, path, body)
	if rr.Code != http.StatusOK {
		s.T().Logf("Status: %d, Body: %s", rr.Code, rr.Body.String())
	}
	require.Equal(s.T(), http.StatusOK, rr.Code)

	var resp app.Install
	require.NoError(s.T(), json.Unmarshal(rr.Body.Bytes(), &resp))
	assert.Equal(s.T(), "updated-name", resp.Name)

	var signalTypes []string
	for _, c := range tests.GetQueueSignals(s.T(), s.deps.DB) {
		signalTypes = append(signalTypes, string(c.Type))
	}
	assert.ElementsMatch(s.T(), []string{"updated", "appconfig-updated"}, signalTypes)
}

func (s *InstallsServiceTestSuite) TestUpdateInstallNotFound() {
	body := map[string]interface{}{
		"name": "updated-name",
	}

	rr := s.makeRequest(http.MethodPatch, "/v1/installs/ins_nonexistent_00000000", body)
	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}

func (s *InstallsServiceTestSuite) TestUpdateInstallApprovalOption() {
	install := s.createTestInstall()

	body := map[string]interface{}{
		"install_config": map[string]interface{}{
			"approval_option": app.InstallApprovalOptionApproveAll,
		},
	}
	rr := s.makeRequest(http.MethodPatch, fmt.Sprintf("/v1/installs/%s", install.ID), body)
	require.Equal(s.T(), http.StatusOK, rr.Code, "body: %s", rr.Body.String())

	var cfg app.InstallConfig
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).
		Where(app.InstallConfig{InstallID: install.ID}).
		Order("created_at DESC").
		First(&cfg).Error)
	assert.Equal(s.T(), app.InstallApprovalOptionApproveAll, cfg.ApprovalOption)
	assert.Equal(s.T(), s.testOrg.ID, cfg.OrgID)
}

// TestUpdateInstallForeignOrgDenied exercises the handler without the installauth
// middleware (tests.NewTestRouter does not mount it), so the handler itself must
// refuse to touch an install owned by another org.
func (s *InstallsServiceTestSuite) TestUpdateInstallForeignOrgDenied() {
	victim := s.createTestInstall()

	ctx2, acc2 := s.deps.Seeder.EnsureAccount(context.Background(), s.T())
	ctx2, org2 := s.deps.Seeder.EnsureOrg(ctx2, s.T())
	seedOrgQueues(s.T(), s.deps.DB, org2)
	app2 := s.deps.Seeder.CreateApp(ctx2, s.T())
	s.deps.Seeder.CreateAppConfig(ctx2, s.T(), app2.ID)
	own := s.deps.Seeder.CreateInstall(ctx2, s.T(), app2)
	seedInstallQueues(s.T(), s.deps.DB, own)

	router2 := tests.NewTestRouter(tests.RouterOptions{
		L:       s.deps.L,
		DB:      s.deps.DB,
		TestOrg: org2,
		TestAcc: acc2,
	})
	require.NoError(s.T(), s.installsService.RegisterPublicRoutes(router2))

	body := map[string]interface{}{
		"name": "pwned",
		"install_config": map[string]interface{}{
			"approval_option": app.InstallApprovalOptionApproveAll,
		},
	}

	// positive control: org2 can update its own install, so RBAC is not what denies below
	rr := patchVia(s, router2, fmt.Sprintf("/v1/installs/%s", own.ID), body)
	require.Equal(s.T(), http.StatusOK, rr.Code, "body: %s", rr.Body.String())
	tests.ClearQueueSignals(s.T(), s.deps.DB)

	rr = patchVia(s, router2, fmt.Sprintf("/v1/installs/%s", victim.ID), body)
	require.Equal(s.T(), http.StatusNotFound, rr.Code, "body: %s", rr.Body.String())
	assert.NotContains(s.T(), rr.Body.String(), victim.Name)

	var after app.Install
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Where(app.Install{ID: victim.ID}).First(&after).Error)
	assert.Equal(s.T(), victim.Name, after.Name)

	var cfgCount int64
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&app.InstallConfig{}).
		Where(app.InstallConfig{InstallID: victim.ID}).Count(&cfgCount).Error)
	assert.Zero(s.T(), cfgCount, "foreign org wrote an install config for the victim install")

	assert.Empty(s.T(), tests.GetQueueSignals(s.T(), s.deps.DB))
}

// TestUpdateInstallForeignConfigRowNotRewritten guards the upsert itself: even if a
// config row for the install is owned by another org, it must not be rewritten.
func (s *InstallsServiceTestSuite) TestUpdateInstallForeignConfigRowNotRewritten() {
	install := s.createTestInstall()

	ctx2, _ := s.deps.Seeder.EnsureAccount(context.Background(), s.T())
	_, org2 := s.deps.Seeder.EnsureOrg(ctx2, s.T())

	foreign := app.InstallConfig{
		InstallID:      install.ID,
		ApprovalOption: app.InstallApprovalOptionPrompt,
	}
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Create(&foreign).Error)
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Model(&foreign).
		UpdateColumn("org_id", org2.ID).Error)

	body := map[string]interface{}{
		"install_config": map[string]interface{}{
			"approval_option": app.InstallApprovalOptionApproveAll,
		},
	}
	rr := s.makeRequest(http.MethodPatch, fmt.Sprintf("/v1/installs/%s", install.ID), body)
	require.Equal(s.T(), http.StatusNotFound, rr.Code, "body: %s", rr.Body.String())

	var cfg app.InstallConfig
	require.NoError(s.T(), s.deps.DB.WithContext(s.ctx).Where(app.InstallConfig{ID: foreign.ID}).First(&cfg).Error)
	assert.Equal(s.T(), app.InstallApprovalOptionPrompt, cfg.ApprovalOption)
}

func patchVia(s *InstallsServiceTestSuite, router *gin.Engine, path string, body interface{}) *httptest.ResponseRecorder {
	b, err := json.Marshal(body)
	require.NoError(s.T(), err)
	req, err := http.NewRequest(http.MethodPatch, path, bytes.NewBuffer(b))
	require.NoError(s.T(), err)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	return rr
}
