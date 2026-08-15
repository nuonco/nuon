package workflow

import (
	"context"
	"os"

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

	cloudEnv, err := h.cloudCredentialEnv(ctx)
	if err != nil {
		return nil, err
	}

	return generics.MergeMap(env, cloudEnv), nil
}

// getContainerBuiltInEnv is the image-backed-action variant of getBuiltInEnv:
// it maps every host path (outputs file, workspace root, kubeconfig) through
// mapPath so the values resolve inside the container's workspace mount
func (h *handler) getContainerBuiltInEnv(ctx context.Context, cfg *models.AppActionWorkflowStepConfig, mapPath func(string) string) (map[string]string, error) {
	env := map[string]string{
		outputsEnvVar: mapPath(h.outputsFP(cfg)),
		rootEnvVar:    mapPath(h.state.workspace.Root()),
	}

	if h.state.plan.ClusterInfo != nil {
		path := h.state.workspace.AbsPath(config.DefaultKubeConfigFilename)
		// unlink any symlink a prior action container may have planted here
		// before writing the kubeconfig into the shared workspace.
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return nil, errors.Wrap(err, "unable to clear existing kubeconfig path")
		}
		if err := config.WriteConfig(ctx, h.state.plan.ClusterInfo, path); err != nil {
			return nil, errors.Wrap(err, "unable to write kube config")
		}
		// client-go writes this 0600, which a container running as a non-root
		// user cannot read. The file holds an exec credential plugin reference
		// rather than a token, and the workspace is already shared with the
		// container, so widening it does not expose anything new.
		if err := os.Chmod(path, 0o644); err != nil {
			return nil, errors.Wrap(err, "unable to make kube config readable by the container")
		}

		env[config.DefaultKubeConfigEnvVar] = mapPath(path)
		env[hasKubeConfigEnvVar] = "true"
	} else {
		env[hasKubeConfigEnvVar] = "false"
	}

	cloudEnv, err := h.cloudCredentialEnv(ctx)
	if err != nil {
		return nil, err
	}

	return generics.MergeMap(env, cloudEnv), nil
}

func (h *handler) cloudCredentialEnv(ctx context.Context) (map[string]string, error) {
	env := map[string]string{}

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
