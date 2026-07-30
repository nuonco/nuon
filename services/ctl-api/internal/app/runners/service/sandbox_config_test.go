package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestSandboxRunnerJobConfigAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`CREATE TABLE sandbox_mode_job_configs (
		id text PRIMARY KEY,
		created_by_id text NOT NULL DEFAULT '',
		created_at datetime NOT NULL,
		updated_at datetime NOT NULL,
		deleted_at integer DEFAULT 0,
		job_type text NOT NULL,
		operation text DEFAULT '',
		enabled numeric DEFAULT true,
		duration integer,
		sleep_duration integer,
		should_error numeric,
		panic numeric,
		trigger_shutdown numeric,
		error_message text DEFAULT '',
		log_template text,
		plan_template text,
		plan_display_template text,
		state_template text,
		output_template text,
		UNIQUE(job_type, operation, deleted_at)
	)`).Error)

	svc := &service{db: db}
	router := gin.New()
	router.PUT("/v1/sandbox-mode/runner-jobs/:job_type", svc.AdminUpsertSandboxConfig)
	router.GET("/v1/runners/:runner_id/sandbox-config", svc.GetRunnerSandboxConfig)

	body := bytes.NewBufferString(`{
		"operation":"create-apply-plan",
		"enabled":true,
		"should_error":true,
		"error_message":"injected runner job failure"
	}`)
	req := httptest.NewRequest(http.MethodPut, "/v1/sandbox-mode/runner-jobs/sandbox-terraform", body)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	require.Equal(t, http.StatusOK, resp.Code)

	var cfg app.SandboxModeJobConfig
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &cfg))
	assert.Equal(t, "sandbox-terraform", cfg.JobType)
	assert.Equal(t, "create-apply-plan", cfg.Operation)
	assert.True(t, cfg.ShouldError)
	assert.Equal(t, "injected runner job failure", cfg.ErrorMessage)

	req = httptest.NewRequest(http.MethodGet, "/v1/runners/runner-id/sandbox-config?job_type=sandbox-terraform&operation=create-apply-plan", nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	require.Equal(t, http.StatusOK, resp.Code)
	assert.Contains(t, resp.Body.String(), "injected runner job failure")

	require.NoError(t, db.Model(&app.SandboxModeJobConfig{}).Where(app.SandboxModeJobConfig{ID: cfg.ID}).Update("enabled", false).Error)
	req = httptest.NewRequest(http.MethodGet, "/v1/runners/runner-id/sandbox-config?job_type=sandbox-terraform&operation=create-apply-plan", nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}
