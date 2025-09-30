package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/patcher"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

type UpdateRunnerSettingsRequest struct {
	ContainerImageURL string `json:"container_image_url"`
	ContainerImageTag string `json:"container_image_tag"`
	RunnerAPIURL      string `json:"runner_api_url"`

	K8sServiceAccountName string `json:"org_k8s_service_account_name"`
	AWSIAMRoleARN         string `json:"org_awsiam_role_arn"`

	AWSMaxInstanceLifetime *int `json:"aws_max_instance_lifetime,omitempty" validate:"omitempty,min=86400,max=31536000"`
}

func (c *UpdateRunnerSettingsRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
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
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
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
		OrgK8sServiceAccountName: req.K8sServiceAccountName,
		OrgAWSIAMRoleARN:         req.AWSIAMRoleARN,
	}
	if req.AWSMaxInstanceLifetime != nil {
		updates.AWSMaxInstanceLifetime = *req.AWSMaxInstanceLifetime
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

	return &obj, nil
}
