package workflow

import (
	"context"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/aws/credentials"
	gcpcredentials "github.com/nuonco/nuon/pkg/gcp/credentials"
	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/kube/config"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

const (
	outputsEnvVar       string = "NUON_ACTIONS_OUTPUT_FILEPATH"
	rootEnvVar                 = "NUON_ACTIONS_ROOT"
	hasKubeConfigEnvVar string = "NUON_KUBECONFIG_ENABLED"
)

func (h *handler) getBuiltInEnv(ctx context.Context, cfg *models.AppActionWorkflowStepConfig) (map[string]string, error) {
	outputsFP := h.outputsFP(cfg)
	env := map[string]string{
		outputsEnvVar: outputsFP,
		rootEnvVar:    h.state.workspace.Root(),
	}

	if h.state.plan.ClusterInfo != nil {
		path := h.state.workspace.AbsPath(config.DefaultKubeConfigFilename)
		if err := config.WriteConfig(ctx, h.state.plan.ClusterInfo, path); err != nil {
			return nil, errors.Wrap(err, "unable to write kube config")
		}

		env[config.DefaultKubeConfigEnvVar] = path
		env[hasKubeConfigEnvVar] = "true"
	} else {
		env[hasKubeConfigEnvVar] = "false"
	}

	if h.state.auth.AWSAuth != nil {
		awsEnv, err := credentials.FetchEnv(ctx, h.state.auth.AWSAuth)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get AWS credentials")
		}

		env = generics.MergeMap(env, awsEnv)
	}

	if h.state.auth.GCPAuth != nil {
		gcpEnv, err := gcpcredentials.FetchEnv(ctx, h.state.auth.GCPAuth)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get GCP credentials")
		}

		if h.state.auth.GCPAuth.ImpersonateServiceAccount != "" {
			gcpEnv["CLOUDSDK_AUTH_IMPERSONATE_SERVICE_ACCOUNT"] = h.state.auth.GCPAuth.ImpersonateServiceAccount
		}

		env = generics.MergeMap(env, gcpEnv)
	}

	return env, nil
}
