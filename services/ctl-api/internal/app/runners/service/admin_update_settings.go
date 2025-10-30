package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	validatorPkg "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/validator"
)

type AdminUpdateRunnerSettingsRequest struct {
	ContainerImageURL string `json:"container_image_url"`
	ContainerImageTag string `json:"container_image_tag"`
	RunnerAPIURL      string `json:"runner_api_url"`

	K8sServiceAccountName string `json:"k8s_service_account_name"`
	AWSIAMRoleARN         string `json:"aws_iam_role_arn"`

	AWSMaxInstanceLifetime int `json:"aws_max_instance_lifetime" default:"604800" validate:"min=86400,max=31536000"`
}

func (c *AdminUpdateRunnerSettingsRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						AdminUpdateRunnerSettings
// @Summary				update a runner's settings
// @Description.markdown	update_runner_settings.md
// @Param					runner_id	path	string								true	"runner ID"
// @Param					req			body	AdminUpdateRunnerSettingsRequest	true	"Input"
// @Tags					runners/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{object}	app.RunnerGroupSettings
// @Router					/v1/runners/{runner_id}/settings [PATCH]
func (s *service) AdminUpdateRunnerSettings(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")

	var req AdminUpdateRunnerSettingsRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	if err := req.Validate(validator.New()); err != nil {
		ctx.Error(err)
		return
	}

	settings, err := s.adminUpdateRunnerSettings(ctx, runnerID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to update settings: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, settings)
}

func (s *service) adminUpdateRunnerSettings(ctx context.Context, runnerID string, req *AdminUpdateRunnerSettingsRequest) (*app.RunnerGroupSettings, error) {
	runner, err := s.getRunner(ctx, runnerID)
	if err != nil {
		return nil, fmt.Errorf("unable to get runner: %w", err)
	}

	updates := app.RunnerGroupSettings{
		ContainerImageURL:        req.ContainerImageURL,
		ContainerImageTag:        req.ContainerImageTag,
		RunnerAPIURL:             req.RunnerAPIURL,
		OrgK8sServiceAccountName: req.K8sServiceAccountName,
		OrgAWSIAMRoleARN:         req.AWSIAMRoleARN,
		AWSMaxInstanceLifetime:   req.AWSMaxInstanceLifetime,
	}
	obj := app.RunnerGroupSettings{
		RunnerGroupID: runner.RunnerGroupID,
	}

	if res := s.db.WithContext(ctx).
		Where(obj).
		Updates(updates); res.Error != nil {
		return nil, fmt.Errorf("unable to update runner settings: %w", res.Error)
	}

	return &obj, nil
}
