package plan

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/principal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	operationroles "github.com/nuonco/nuon/services/ctl-api/internal/pkg/operation-roles"
)

func (p *Planner) createSyncPlan(ctx workflow.Context, req *CreateSyncPlanRequest) (*plantypes.SyncOCIPlan, error) {
	deploy, err := activities.AwaitGetDeployByDeployID(ctx, req.InstallDeployID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install deploy")
	}

	srcCfg, err := p.getOrgRegistryRepositoryConfig(ctx, req.InstallID, req.InstallDeployID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get org registry repository")
	}

	dstCfg, err := p.getInstallRegistryRepositoryConfig(ctx, req.InstallID, req.InstallDeployID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install registry repository")
	}

	// Get context for role selection
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, err
	}

	install, err := activities.AwaitGetByInstallID(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config")
	}

	stack, err := activities.AwaitGetInstallStackByInstallID(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install stack")
	}

	installState, err := activities.AwaitGetInstallState(ctx, &activities.GetInstallStateRequest{
		InstallID: install.ID,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install state")
	}

	// Perform role selection for OCI sync
	// OCI sync is part of the deployment process, so use OperationDeploy
	operation := app.OperationDeploy
	defaultRole := appCfg.PermissionsConfig.ProvisionRole.Name

	selectionCtx := &operationroles.SelectionContext{
		Operation:     operation,
		PrincipalType: principal.TypeComponent,
		PrincipalName: deploy.ComponentName,
		RuntimeRole:   deploy.Role,
		EntityRoles:   nil, // OCI sync doesn't have component-specific operation roles
		MatrixRules:   appCfg.OperationRoleConfig.Rules,
		DefaultRole:   defaultRole,
		AppConfig:     appCfg,
		StackOutputs:  &stack.InstallStackOutputs,
		InstallState:  installState,
	}

	roleSelection, err := operationroles.SelectRole(selectionCtx, l)
	if err != nil {
		l.Warn("dynamic role selection failed, falling back to default role",
			zap.Error(err),
			zap.String("default_role", selectionCtx.DefaultRole),
		)

		var fallbackErr error
		roleSelection, fallbackErr = operationroles.GetDefaultRoleSelection(selectionCtx)
		if fallbackErr != nil {
			return nil, fmt.Errorf("unable to get default role: %w", fallbackErr)
		}

		l.Warn("using default role for OCI sync",
			zap.String("role_name", roleSelection.RoleName),
			zap.String("role_arn", roleSelection.RoleARN),
		)
	}

	l.Info("selected role for OCI sync plan",
		zap.String("role_name", roleSelection.RoleName),
		zap.String("role_arn", roleSelection.RoleARN),
		zap.String("source", string(roleSelection.Source)),
		zap.String("operation", string(operation)),
		zap.String("component_name", deploy.ComponentName),
	)

	// Create auth configuration using selected role
	var awsAuth *awscredentials.Config
	var azureAuth *azurecredentials.Config

	switch {
	case stack.InstallStackOutputs.AWSStackOutputs != nil:
		awsAuth = &awscredentials.Config{
			Region: stack.InstallStackOutputs.AWSStackOutputs.Region,
			AssumeRole: &awscredentials.AssumeRoleConfig{
				SessionName: fmt.Sprintf("oci-sync-%s", deploy.ID),
				RoleARN:     roleSelection.RoleARN,
			},
		}
	case stack.InstallStackOutputs.AzureStackOutputs != nil:
		azureOutputs := stack.InstallStackOutputs.AzureStackOutputs
		azureAuth = &azurecredentials.Config{
			ServicePrincipal: &azurecredentials.ServicePrincipalCredentials{
				SubscriptionID:       azureOutputs.SubscriptionID,
				SubscriptionTenantID: azureOutputs.SubscriptionTenantID,
			},
			UseDefault: true,
		}
	}

	pln := &plantypes.SyncOCIPlan{
		Src:    srcCfg,
		SrcTag: deploy.ComponentBuildID,

		DstTag: deploy.ID,
		Dst:    dstCfg,

		AWSAuth:   awsAuth,
		AzureAuth: azureAuth,
	}

	org, err := activities.AwaitGetOrgByInstallID(ctx, deploy.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get org")
	}
	if org.SandboxMode {
		pln.SandboxMode = &plantypes.SandboxMode{
			Enabled: true,
			Outputs: map[string]any{
				"image": map[string]interface{}{
					"tag":           "v1.2.3",
					"repository":    "nuon/app-service",
					"media_type":    "application/vnd.docker.distribution.manifest.v2+json",
					"digest":        "sha256:a123b456c789d012e345f678g901h234i567j890k123l456m789n012o345p",
					"size":          28437192,
					"urls":          []string{"registry.example.com/nuon/app-service:v1.2.3"},
					"annotations":   map[string]string{"org.opencontainers.image.created": "2024-04-29T10:15:30Z"},
					"artifact_type": "application/vnd.docker.container.image.v1+json",
					"platform": map[string]any{
						"architecture": "arm64",
						"os":           "linux",
						"os_version":   "10.0",
						"variant":      "v8",
						"os_features":  []string{"sse4", "aes"},
					},
				},
			},
		}
	}

	return pln, nil
}
