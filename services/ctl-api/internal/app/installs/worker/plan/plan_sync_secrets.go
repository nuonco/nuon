package plan

import (
	"fmt"

	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	awscredentials "github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/generics"
	plantypes "github.com/powertoolsdev/mono/pkg/plans/types"
	"github.com/powertoolsdev/mono/pkg/render"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
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

		secret, err := p.getKubernetesSecret(stack.InstallStackOutputs, cfg)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get kubernetes secret")
		}
		secrets = append(secrets, secret)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get kubernetes secret")
		}
	}

	if len(secrets) < 1 {
		return &plantypes.SyncSecretsPlan{}, nil
	}

	clusterInfo, err := p.getKubeClusterInfo(ctx, stack, state)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get cluster information")
	}

	plan := &plantypes.SyncSecretsPlan{
		ClusterInfo: clusterInfo,
		AWSAuth: &awscredentials.Config{
			Region: stack.InstallStackOutputs.AWSStackOutputs.Region,
			AssumeRole: &awscredentials.AssumeRoleConfig{
				SessionName: fmt.Sprintf("install-sync-secrets-%s", req.InstallID),
				RoleARN:     stack.InstallStackOutputs.AWSStackOutputs.ProvisionIAMRoleARN,
			},
		},
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

func (p *Planner) getKubernetesSecret(stack app.InstallStackOutputs, cfg app.AppSecretConfig) (plantypes.KubernetesSecretSync, error) {
	key := fmt.Sprintf("%s_arn", cfg.Name)
	secretARN, ok := stack.Data[key]
	if !ok {
		return plantypes.KubernetesSecretSync{}, fmt.Errorf("secret arn not found in stack output: %s", key)
	}

	return plantypes.KubernetesSecretSync{
		SecretARN:  generics.FromPtrStr(secretARN),
		SecretName: cfg.Name,

		Namespace: cfg.KubernetesSecretNamespace,
		Name:      cfg.KubernetesSecretName,
		KeyName:   cfg.KubernetesSecretKey,
		Format:    string(cfg.Format),
	}, nil
}
