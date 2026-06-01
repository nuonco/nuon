package plan

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	azurecredentials "github.com/nuonco/nuon/pkg/azure/credentials"
	gcpcredentials "github.com/nuonco/nuon/pkg/gcp/credentials"
	"github.com/nuonco/nuon/pkg/generics"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
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
		// v2 targets imply sync is enabled even when the legacy kubernetes_sync flag is unset/false.
		if !cfg.KubernetesSync && len(cfg.KubernetesSyncTargets) == 0 {
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

	// Build cloud auth based on the cloud provider
	var cloudAuth *CloudAuth
	switch {
	case stack.InstallStackOutputs.AWSStackOutputs != nil:
		if stack.InstallStackOutputs.AWSStackOutputs.ProvisionIAMRoleARN == "" {
			return nil, fmt.Errorf("provision role not enabled in install stack")
		}
		cloudAuth = &CloudAuth{
			AWS: &awscredentials.Config{
				Region: stack.InstallStackOutputs.AWSStackOutputs.Region,
				AssumeRole: &awscredentials.AssumeRoleConfig{
					SessionName: fmt.Sprintf("install-sync-secrets-%s", req.InstallID),
					RoleARN:     stack.InstallStackOutputs.AWSStackOutputs.ProvisionIAMRoleARN,
				},
			},
		}
	case stack.InstallStackOutputs.AzureStackOutputs != nil:
		azureOutputs := stack.InstallStackOutputs.AzureStackOutputs
		cloudAuth = &CloudAuth{
			Azure: &azurecredentials.Config{
				ServicePrincipal: &azurecredentials.ServicePrincipalCredentials{
					SubscriptionID:       azureOutputs.SubscriptionID,
					SubscriptionTenantID: azureOutputs.SubscriptionTenantID,
				},
				UseDefault: true,
			},
		}
	case stack.InstallStackOutputs.GCPStackOutputs != nil:
		gcpOutputs := stack.InstallStackOutputs.GCPStackOutputs
		if gcpOutputs.ProvisionSAEmail == "" {
			return nil, fmt.Errorf("provision service account not enabled in install stack")
		}
		cloudAuth = &CloudAuth{
			GCP: &gcpcredentials.Config{
				ProjectID:                 gcpOutputs.ProjectID,
				Region:                    gcpOutputs.Region,
				ImpersonateServiceAccount: gcpOutputs.ProvisionSAEmail,
			},
		}
	default:
		return nil, errors.New("secret sync not supported on current cloud provider")
	}

	clusterInfo, err := p.getKubeClusterInfo(ctx, stack, state, cloudAuth)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cluster information")
	}

	plan := &plantypes.SyncSecretsPlan{
		ClusterInfo:       clusterInfo,
		AWSAuth:           cloudAuth.AWS,
		AzureAuth:         cloudAuth.Azure,
		GCPAuth:           cloudAuth.GCP,
		KubernetesSecrets: secrets,
	}

	if install.SandboxMode.Bool {
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
	sync := plantypes.KubernetesSecretSync{
		SecretName: cfg.Name,
		Namespace:  cfg.KubernetesSecretNamespace,
		Name:       cfg.KubernetesSecretName,
		KeyName:    cfg.KubernetesSecretKey,
		Format:     string(cfg.Format),
		Targets:    kubernetesSyncTargets(cfg.KubernetesSyncTargets),
	}

	switch {
	case stack.GCPStackOutputs != nil:
		key := fmt.Sprintf("%s_secret_name", cfg.Name)
		val, ok := stack.Data[key]
		if !ok || val == nil || generics.FromPtrStr(val) == "" {
			if cfg.Required {
				return plantypes.KubernetesSecretSync{}, false, fmt.Errorf("secret name not found in stack output: %s", key)
			}
			return plantypes.KubernetesSecretSync{}, false, nil
		}
		sync.GCPSecretName = generics.FromPtrStr(val)
	case stack.AzureStackOutputs != nil:
		key := fmt.Sprintf("%s_secret_id", cfg.Name)
		val, ok := stack.Data[key]
		if !ok || val == nil || generics.FromPtrStr(val) == "" {
			if cfg.Required {
				return plantypes.KubernetesSecretSync{}, false, fmt.Errorf("secret id not found in stack output: %s", key)
			}
			return plantypes.KubernetesSecretSync{}, false, nil
		}
		sync.AzureKeyVaultSecretID = generics.FromPtrStr(val)
	default:
		key := fmt.Sprintf("%s_arn", cfg.Name)
		val, ok := stack.Data[key]
		if !ok || val == nil || generics.FromPtrStr(val) == "" {
			if cfg.Required {
				return plantypes.KubernetesSecretSync{}, false, fmt.Errorf("secret arn not found in stack output: %s", key)
			}
			return plantypes.KubernetesSecretSync{}, false, nil
		}
		sync.SecretARN = generics.FromPtrStr(val)
	}

	return sync, true, nil
}

// kubernetesSyncTargets maps the model's v2 sync targets onto the plan type the runner consumes. Returns nil when no
// targets are configured so the runner falls back to the v1 single-destination path.
func kubernetesSyncTargets(targets []app.AppSecretKubernetesSyncTarget) []plantypes.KubernetesSecretSyncTarget {
	if len(targets) == 0 {
		return nil
	}

	out := make([]plantypes.KubernetesSecretSyncTarget, 0, len(targets))
	for _, t := range targets {
		out = append(out, plantypes.KubernetesSecretSyncTarget{
			Namespaces: []string(t.Namespaces),
			Name:       t.Name,
			Key:        t.Key,
		})
	}

	return out
}
