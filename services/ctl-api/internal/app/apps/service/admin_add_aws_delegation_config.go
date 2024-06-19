package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type AdminAddAWSDelegationRequest struct {
	IAMRoleARN string `json:"iam_role_arn"`

	GovCloud        bool   `json:"gov_cloud"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
}

// @ID AdminAddAWSDelegationConfig
// @Summary	add a delegation config to an app
// @Description.markdown admin_add_aws_delegation.md
// @Tags			apps/admin
// @Accept			json
// @Param			req		body	AdminAddAWSDelegationRequest	true	"Input"
// @Param			app_id	path	string					true	"app id"
// @Produce		json
// @Success		201	{string}	ok
// @Router			/v1/apps/{app_id}/admin-aws-delegation [POST]
func (s *service) AdminAddAWSDelegationConfig(ctx *gin.Context) {
	appID := ctx.Param("app_id")

	var req AdminAddAWSDelegationRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	currentApp, err := s.getApp(ctx, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app: %w", err))
		return
	}

	if len(currentApp.AppSandboxConfigs) < 1 {
		ctx.Error(fmt.Errorf("at least one app sandbox config must be synced first"))
		return
	}

	if err := s.adminAddAWSDelegationConfig(ctx, currentApp, &req); err != nil {
		ctx.Error(fmt.Errorf("unable to create aws delegation config: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, true)
}

func (s *service) adminAddAWSDelegationConfig(ctx context.Context, currentApp *app.App, req *AdminAddAWSDelegationRequest) error {
	if req.GovCloud {
		res := s.db.Model(app.AppSandboxConfig{
			ID: currentApp.AppSandboxConfig.ID,
		}).Updates(app.AppSandboxConfig{
			AWSRegionType: generics.NewNullString(app.AWSRegionTypeGovCloud.String()),
		})
		if res.Error != nil {
			return fmt.Errorf("unable to update app sandbox to gov cloud: %w", res.Error)
		}
	}

	cfg := app.AppAWSDelegationConfig{
		CreatedByID:        currentApp.AppSandboxConfig.CreatedByID,
		OrgID:              currentApp.AppSandboxConfig.OrgID,
		AppSandboxConfigID: currentApp.AppSandboxConfig.ID,

		// fields
		IAMRoleARN:      req.IAMRoleARN,
		AccessKeyID:     req.AccessKeyID,
		SecretAccessKey: req.SecretAccessKey,
	}
	res := s.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			UpdateAll: true,
			Columns: []clause.Column{
				{
					Name: "deleted_at",
				},
				{
					Name: "app_sandbox_config_id",
				},
			},
		}).
		Create(&cfg)
	if res.Error != nil {
		return fmt.Errorf("unable to create app delegation config: %w", res.Error)
	}

	return nil
}
