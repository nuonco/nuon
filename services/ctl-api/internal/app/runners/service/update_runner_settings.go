package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/restart"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/patcher"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type UpdateRunnerSettingsRequest struct {
	ContainerImageURL  string `json:"container_image_url"`
	ContainerImageTag  string `json:"container_image_tag"`
	RunnerAPIURL       string `json:"runner_api_url"`
	BinaryVersion      string `json:"binary_version"`
	ContainerMaxUptime int    `json:"container_max_uptime"`
	VMMaxUptime        int    `json:"vm_max_uptime"`

	K8sServiceAccountName string `json:"org_k8s_service_account_name"`
	AWSIAMRoleARN         string `json:"org_awsiam_role_arn"`

	AWSAuthMethod app.RunnerAWSAuthMethod `json:"aws_auth_method,omitempty" validate:"omitempty,oneof=iid sts" swaggertype:"string" enums:"iid,sts"`

	// Deprecated: no longer used. Instance refresh is handled by a backend cron.
	AWSMaxInstanceLifetime *int `json:"aws_max_instance_lifetime,omitempty" validate:"omitempty,min=86400,max=31536000"`

	// JobGroupParallelism maps job group names to max-in-flight values for parallel job execution.
	// e.g., {"build": 2, "deploy": 1}. Only effective when parallel-runner-jobs feature flag is enabled.
	JobGroupParallelism map[string]int `json:"job_group_parallelism,omitempty"`
}

func (c *UpdateRunnerSettingsRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						UpdateRunnerSettings
// @Summary				update a runner's settings via its runner settings group
// @Description.markdown	update_runner_settings.md
// @Param					req						body	UpdateRunnerSettingsRequest	true	"Input"
// @Param					runner_id			path	string							true	"runner ID"
// @Tags					runners
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.RunnerJobExecution
// @Router					/v1/runners/{runner_id}/settings [PATCH]
func (s *service) UpdateRunnerSettings(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	runnerID := ctx.Param("runner_id")

	var req UpdateRunnerSettingsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	if err := req.Validate(validator.New()); err != nil {
		ctx.Error(err)
		return
	}

	settings, err := s.updateRunnerSettings(ctx, runnerID, org.ID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to update settings: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, settings)
}

func (s *service) updateRunnerSettings(ctx context.Context, runnerID, orgID string, req *UpdateRunnerSettingsRequest) (*app.RunnerGroupSettings, error) {
	runner, err := s.getOrgRunner(ctx, runnerID, orgID)
	if err != nil {
		return nil, fmt.Errorf("unable to get runner: %w", err)
	}

	if runner.RunnerGroup.ID == "" {
		return nil, fmt.Errorf("runner %s does not have a runner group", runnerID)
	}

	updates := app.RunnerGroupSettings{
		ContainerImageURL:        req.ContainerImageURL,
		ContainerImageTag:        req.ContainerImageTag,
		RunnerAPIURL:             req.RunnerAPIURL,
		BinaryVersion:            req.BinaryVersion,
		ContainerMaxUptime:       req.ContainerMaxUptime,
		VMMaxUptime:              req.VMMaxUptime,
		OrgK8sServiceAccountName: req.K8sServiceAccountName,
		OrgAWSIAMRoleARN:         req.AWSIAMRoleARN,
		AWSAuthMethod:            req.AWSAuthMethod,
	}
	if req.AWSMaxInstanceLifetime != nil {
		updates.AWSMaxInstanceLifetime = *req.AWSMaxInstanceLifetime
	}
	if len(req.JobGroupParallelism) > 0 {
		h := pgtype.Hstore{}
		for k, v := range req.JobGroupParallelism {
			s := strconv.Itoa(v)
			h[k] = &s
		}
		updates.JobGroupParallelism = h
	}

	obj := app.RunnerGroupSettings{
		RunnerGroupID: runner.RunnerGroupID,
	}

	if res := s.db.WithContext(ctx).
		Scopes(scopes.WithPatcher(patcher.PatcherOptions{})).
		Where(obj).
		UpdateColumns(updates); res.Error != nil {
		return nil, fmt.Errorf("unable to update runner settings: %w", res.Error)
	}

	if req.ContainerImageTag != "" {
		if err := s.helpers.EnqueueRunnerSignal(ctx, runnerID, &restart.Signal{RunnerID: runnerID}); err != nil {
			return nil, fmt.Errorf("unable to enqueue restart signal: %w", err)
		}
	}

	return &obj, nil
}
