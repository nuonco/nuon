package helpers

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	pkggenerics "github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type InstallMetadata struct {
	ManagedBy string `json:"managed_by,omitempty"`
}

type CreateInstallParams struct {
	Name string `json:"name" validate:"required"`

	AWSAccount *struct {
		Region string `json:"region"`
	} `json:"aws_account"`

	AzureAccount *struct {
		Location string `json:"location"`
	} `json:"azure_account"`

	GCPAccount *struct {
		ProjectID string `json:"project_id"`
		Region    string `json:"region"`
	} `json:"gcp_account"`

	Inputs map[string]*string `json:"inputs"`

	InstallConfig *CreateInstallConfigParams `json:"install_config"`

	Metadata InstallMetadata `json:"metadata,omitempty"`

	SandboxMode bool `json:"sandbox_mode,omitempty" swaggerignore:"true"`
}

func (s *Helpers) CreateInstall(ctx context.Context, appID string, req *CreateInstallParams) (*app.Install, error) {
	parentApp := app.App{}
	res := s.db.WithContext(ctx).
		Preload("Components").
		Preload("AppSandboxConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_sandbox_configs.created_at DESC").Limit(1)
		}).
		Preload("AppRunnerConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_runner_configs.created_at DESC").Limit(1)
		}).
		Preload("AppInputConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_input_configs.created_at DESC").Limit(1)
		}).
		Preload("AppConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order(views.TableOrViewName(s.db, &app.AppConfig{}, ".created_at DESC")).Limit(1)
		}).
		Preload("AppPermissionsConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order("app_permissions_configs.created_at DESC").Limit(1)
		}).
		Preload("AppPermissionsConfigs.Roles").
		First(&parentApp, "id = ?", appID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app: %w", res.Error)
	}

	// make sure the inputs are valid
	latestAppInputConfig, err := s.GetLatestAppInputConfig(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("unable to get latest app input config: %w", err)
	}
	if err := s.ValidateInstallInputs(ctx, latestAppInputConfig, req.Inputs); err != nil {
		return nil, err
	}

	install := app.Install{
		AppID:              appID,
		Name:               req.Name,
		SandboxMode:        pkggenerics.NewNullBool(req.SandboxMode),
		AppSandboxConfigID: parentApp.AppSandboxConfigs[0].ID,
		AppRunnerConfigID:  parentApp.AppRunnerConfigs[0].ID,
		AppConfigID:        parentApp.AppConfigs[0].ID,
		InstallSandbox: app.InstallSandbox{
			Status: app.InstallSandboxStatusQueued,
			TerraformWorkspace: app.TerraformWorkspace{
				ID: domains.NewTerraformWorkspaceID(),
			},
		},
		Metadata: generics.ToHstore(map[string]string{
			"managed_by": req.Metadata.ManagedBy,
		}),
	}

	if req.AWSAccount != nil {
		install.AWSAccount = &app.AWSAccount{
			Region: req.AWSAccount.Region,
		}
	}
	if req.AzureAccount != nil {
		install.AzureAccount = &app.AzureAccount{
			Location: req.AzureAccount.Location,
		}
	}
	if req.GCPAccount != nil {
		install.GCPAccount = &app.GCPAccount{
			ProjectID: req.GCPAccount.ProjectID,
			Region:    req.GCPAccount.Region,
		}
	}
	if parentApp.AppPermissionsConfig.ID != "" && len(parentApp.AppPermissionsConfig.Roles) > 0 {
		installRoles := make([]app.InstallRoles, 0)

		for _, role := range parentApp.AppPermissionsConfig.Roles {
			installRoles = append(installRoles, app.InstallRoles{
				AppRoleConfigID: role.ID,
			})
		}

		install.InstallRoles = installRoles
	}

	if len(parentApp.AppInputConfigs) > 0 {
		install.InstallInputs = []app.InstallInputs{
			{
				Values:           req.Inputs,
				AppInputConfigID: parentApp.AppInputConfig.ID,
			},
		}
	}

	switch parentApp.AppRunnerConfigs[0].Type {
	case "aws":
		install.InstallStack = &app.InstallStack{
			InstallStackOutputs: app.InstallStackOutputs{
				Data: generics.ToHstore(map[string]string{}),
			},
		}
	case "azure":
		install.InstallStack = &app.InstallStack{
			InstallStackOutputs: app.InstallStackOutputs{
				Data: generics.ToHstore(map[string]string{}),
			},
		}
	case "gcp":
		install.InstallStack = &app.InstallStack{
			InstallStackOutputs: app.InstallStackOutputs{
				Data: generics.ToHstore(map[string]string{}),
			},
		}
	}

	res = s.db.WithContext(ctx).Create(&install)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to create install: %w", res.Error)
	}

	// Create the install-workflows queue (orchestrates workflow execution, limited concurrency)
	_, err = s.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     install.ID,
		OwnerType:   plugins.TableName(s.db, app.Install{}),
		Namespace:   "installs",
		Name:        InstallWorkflowsQueueName,
		MaxInFlight: 10,
		MaxDepth:    50,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create install-workflows queue: %w", err)
	}

	// Create the install-signals queue (handles individual signal execution, higher concurrency)
	_, err = s.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     install.ID,
		OwnerType:   plugins.TableName(s.db, app.Install{}),
		Namespace:   "installs",
		Name:        InstallSignalsQueueName,
		MaxInFlight: 20,
		MaxDepth:    50,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create install-signals queue: %w", err)
	}

	// Create the install-workflow-steps queue (executes individual workflow steps as signals)
	_, err = s.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OwnerID:     install.ID,
		OwnerType:   plugins.TableName(s.db, app.Install{}),
		Namespace:   "installs",
		Name:        InstallWorkflowStepsQueueName,
		MaxInFlight: 10,
		MaxDepth:    50,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create install-workflow-steps queue: %w", err)
	}

	if req.InstallConfig != nil {
		_, err := s.CreateInstallConfig(ctx, install.ID, req.InstallConfig)
		if err != nil {
			return nil, fmt.Errorf("unable to create install config: %w", err)
		}
	}

	if err := s.componentHelpers.EnsureInstallComponents(ctx, appID, []string{install.ID}); err != nil {
		return nil, fmt.Errorf("unable to ensure install components: %w", err)
	}
	if err := s.actionsHelpers.EnsureInstallAction(ctx, appID, []string{install.ID}); err != nil {
		return nil, fmt.Errorf("unable to ensure install components: %w", err)
	}

	//if err := s.EnsureInstallSandbox(ctx, appID, []string{install.ID}); err != nil {
	//return nil, fmt.Errorf("unable to ensure install components: %w", err)
	//}

	loadedInstall, err := s.getInstall(ctx, install.ID)
	if err != nil {
		return nil, fmt.Errorf("unable to load all install resources: %w", err)
	}

	if _, err := s.runnersHelpers.CreateInstallRunnerGroup(ctx, loadedInstall); err != nil {
		return nil, fmt.Errorf("unable to create install runner: %w", err)
	}

	return &install, nil
}
