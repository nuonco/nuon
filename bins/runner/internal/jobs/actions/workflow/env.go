package workflow

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"
	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/kube/config"
)

const (
	outputsEnvVar string = "NUON_ACTIONS_OUTPUT_FILEPATH"
	rootEnvVar           = "NUON_ACTIONS_ROOT"
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
	}

	if h.state.plan.AWSAuth != nil {
		awsEnv, err := credentials.FetchEnv(ctx, h.state.plan.AWSAuth)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get AWS credentials")
		}

		env = generics.MergeMap(env, awsEnv)
	}

	return env, nil
}
