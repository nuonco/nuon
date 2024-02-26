package helpers

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"gorm.io/gorm"
)

type CreateInstallParams struct {
	Name string `json:"name" validate:"required"`

	AWSAccount struct {
		Region     string `json:"region"`
		IAMRoleARN string `json:"iam_role_arn" validate:"required"`
	} `json:"aws_account" validate:"required"`

	Inputs map[string]*string `json:"inputs"`
}

func (s *Helpers) CreateInstall(ctx context.Context, appID string, req *CreateInstallParams) (*app.Install, error) {
	parentApp := app.App{}
	res := s.db.WithContext(ctx).
		Preload("Components").
		Preload("AppSandboxConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_sandbox_configs.created_at DESC")
		}).
		Preload("AppRunnerConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_runner_configs.created_at DESC")
		}).
		Preload("AppInputConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_input_configs.created_at DESC")
		}).
		First(&parentApp, "id = ?", appID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}
	if len(parentApp.AppSandboxConfigs) < 1 {
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("app does not have any sandbox configs"),
			Description: "please make create at least one app sandbox config first",
		}
	}
	if len(parentApp.AppRunnerConfigs) < 1 {
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("app does not have any runner configs"),
			Description: "please make create at least one app runner config first",
		}
	}

	if err := s.ValidateInstallInputs(ctx, appID, req.Inputs); err != nil {
		return nil, err
	}

	install := app.Install{
		AppID:             appID,
		Name:              req.Name,
		Status:            "queued",
		StatusDescription: "waiting for event loop to start and provision install",
		AWSAccount: app.AWSAccount{
			Region:     req.AWSAccount.Region,
			IAMRoleARN: req.AWSAccount.IAMRoleARN,
		},
		AppSandboxConfigID: parentApp.AppSandboxConfigs[0].ID,
		AppRunnerConfigID:  parentApp.AppRunnerConfigs[0].ID,
	}
	if len(parentApp.AppInputConfigs) > 0 {
		install.InstallInputs = []app.InstallInputs{
			{
				Values:           req.Inputs,
				AppInputConfigID: parentApp.AppInputConfigs[0].ID,
			},
		}
	}

	res = s.db.WithContext(ctx).Create(&install)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create install: %w", res.Error)
	}

	if err := s.componentHelpers.EnsureInstallComponents(ctx, appID, []string{install.ID}); err != nil {
		return nil, fmt.Errorf("unable to ensure install components: %w", err)
	}

	return &install, nil
}
