package config

import (
	"context"

	"github.com/pkg/errors"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"

	"github.com/powertoolsdev/mono/pkg/kube"
)

const (
	DefaultKubeConfigFilename string = ".kubeconfig"
	DefaultKubeConfigEnvVar   string = "KUBECONFIG"
)

func WriteConfig(ctx context.Context, cfg *kube.ClusterInfo, fp string) error {
	kubeCfg, err := kube.ConfigForCluster(ctx, cfg)
	if err != nil {
		return errors.Wrap(err, "unable to get kube config")
	}

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

	if err := clientcmd.WriteToFile(*apiConfig, fp); err != nil {
		return errors.Wrap(err, "unable to write kube config")
	}

	return nil
}
