package plan

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
	"github.com/nuonco/nuon/pkg/generics"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/principal"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	operationroles "github.com/nuonco/nuon/services/ctl-api/internal/pkg/operation-roles"
)

func (p *Planner) createSyncSecretsPlan(ctx workflow.Context, req *CreateSyncSecretsPlanRequest) (*plantypes.SyncSecretsPlan, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return nil, err
	}

	install, err := activities.AwaitGetByInstallID(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install")
	}

	l.Debug("fetching install state")
	state, err := activities.AwaitGetInstallStateByInstallID(ctx, req.InstallID)
	if err != nil {
		l.Error("unable to get install state", zap.Error(err))
		return nil, errors.Wrap(err, "unable to get install state")
	}
	stateData, err := state.WorkflowSafeAsMap(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to generate install map data")
	}

	stack, err := activities.AwaitGetInstallStackByInstallID(ctx, req.InstallID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install stack")
	}

	appCfg, err := activities.AwaitGetAppConfigByID(ctx, install.AppConfigID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get app config")
	}

	if err := render.RenderStruct(&appCfg.SecretsConfig, stateData); err != nil {
		return nil, errors.Wrap(err, "unable to render secrets config")
	}

	secrets := make([]plantypes.KubernetesSecretSync, 0)
	for _, cfg := range appCfg.SecretsConfig.Secrets {
		if !cfg.KubernetesSync {
			continue
		}

		secret, ok, err := p.getKubernetesSecret(stack.InstallStackOutputs, cfg)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get kubernetes secret")
		}
		if !ok {
			l.Debug("skipping optional kubernetes secret sync because stack output is missing or empty", zap.String("secret_name", cfg.Name))
			continue
		}

		secrets = append(secrets, secret)
	}

	if len(secrets) < 1 {
		return &plantypes.SyncSecretsPlan{}, nil
	}

	clusterInfo, err := p.getKubeClusterInfo(ctx, stack, state)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cluster information")
	}

	// Perform role selection for secret sync
	// Secret sync is part of provisioning/deployment, so use OperationDeploy
	operation := app.OperationDeploy
	defaultRole := appCfg.PermissionsConfig.ProvisionRole.Name

	selectionCtx := &operationroles.SelectionContext{
		Operation:     operation,
		PrincipalType: principal.TypeSandbox, // Secrets are synced at install level, use sandbox type
		PrincipalName: "",                    // No specific principal name for secret sync
		RuntimeRole:   "",                    // No runtime role for secret sync
		EntityRoles:   nil,                   // No entity-specific operation roles for secrets
		MatrixRules:   appCfg.OperationRoleConfig.Rules,
		DefaultRole:   defaultRole,
		AppConfig:     appCfg,
		StackOutputs:  &stack.InstallStackOutputs,
		InstallState:  state,
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

		l.Warn("using default role for secret sync",
			zap.String("role_name", roleSelection.RoleName),
			zap.String("role_arn", roleSelection.RoleARN),
		)
	}

	l.Info("selected role for secret sync plan",
		zap.String("role_name", roleSelection.RoleName),
		zap.String("role_arn", roleSelection.RoleARN),
		zap.String("source", string(roleSelection.Source)),
		zap.String("operation", string(operation)),
	)

	// Create auth configuration using selected role
	var awsAuth *awscredentials.Config
	var azureAuth *azurecredentials.Config

	switch {
	case stack.InstallStackOutputs.AWSStackOutputs != nil:
		awsAuth = &awscredentials.Config{
			Region: stack.InstallStackOutputs.AWSStackOutputs.Region,
			AssumeRole: &awscredentials.AssumeRoleConfig{
				SessionName: fmt.Sprintf("install-sync-secrets-%s", req.InstallID),
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

	plan := &plantypes.SyncSecretsPlan{
		ClusterInfo:       clusterInfo,
		AWSAuth:           awsAuth,
		AzureAuth:         azureAuth,
		KubernetesSecrets: secrets,
	}

	org, err := activities.AwaitGetOrgByInstallID(ctx, install.ID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get org")
	}
	if org.SandboxMode {
		plan.SandboxMode = &plantypes.SandboxMode{
			Enabled: true,
			Outputs: map[string]any{
				"TBD": "TBD",
			},
		}
	}

	return plan, nil
}

func (p *Planner) getKubernetesSecret(stack app.InstallStackOutputs, cfg app.AppSecretConfig) (plantypes.KubernetesSecretSync, bool, error) {
	key := fmt.Sprintf("%s_arn", cfg.Name)
	secretARN, ok := stack.Data[key]
	if !ok || secretARN == nil || generics.FromPtrStr(secretARN) == "" {
		if cfg.Required {
			return plantypes.KubernetesSecretSync{}, false, fmt.Errorf("secret arn not found in stack output: %s", key)
		}

		return plantypes.KubernetesSecretSync{}, false, nil
	}

	return plantypes.KubernetesSecretSync{
		SecretARN:  generics.FromPtrStr(secretARN),
		SecretName: cfg.Name,

		Namespace: cfg.KubernetesSecretNamespace,
		Name:      cfg.KubernetesSecretName,
		KeyName:   cfg.KubernetesSecretKey,
		Format:    string(cfg.Format),
	}, true, nil
}
