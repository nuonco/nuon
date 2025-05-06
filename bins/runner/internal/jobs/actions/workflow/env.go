package workflow

import (
	"context"

	"github.com/nuonco/nuon-runner-go/models"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"github.com/powertoolsdev/mono/pkg/aws/credentials"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/kube"
)

const (
	outputsEnvVar      string = "NUON_ACTIONS_OUTPUT_FILEPATH"
	rootEnvVar                = "NUON_ACTIONS_ROOT"
	kubeConfigEnvVars         = "KUBECONFIG"
	kubeConfigFilename        = ".kubeconfig"
)

func (h *handler) getBuiltInEnv(ctx context.Context, cfg *models.AppActionWorkflowStepConfig) (map[string]string, error) {
	outputsFP := h.outputsFP(cfg)
	env := map[string]string{
		outputsEnvVar: outputsFP,
		rootEnvVar:    h.state.workspace.Root(),
	}

	if h.state.plan.ClusterInfo != nil {
		kubeCfg, err := kube.ConfigForCluster(ctx, h.state.plan.ClusterInfo)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get kube config")
		}

		path := h.state.workspace.AbsPath(kubeConfigFilename)

		// Convert rest.Config to clientcmdapi.Config
		apiConfig := &clientcmdapi.Config{
			Kind:           "Config",
			APIVersion:     "v1",
			CurrentContext: "default",
			Clusters: map[string]*clientcmdapi.Cluster{
				"default": {
					Server:                   kubeCfg.Host,
					CertificateAuthorityData: kubeCfg.TLSClientConfig.CAData,
				},
			},
			Contexts: map[string]*clientcmdapi.Context{
				"default": {
					Cluster:  "default",
					AuthInfo: "default",
				},
			},
			AuthInfos: map[string]*clientcmdapi.AuthInfo{
				"default": {
					Exec: kubeCfg.ExecProvider,
				},
			},
		}

		// Write the config to file
		if err := clientcmd.WriteToFile(*apiConfig, path); err != nil {
			return nil, errors.Wrap(err, "unable to write kube config")
		}
		env[kubeConfigEnvVars] = path
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
