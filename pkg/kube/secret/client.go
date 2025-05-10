package secret

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/powertoolsdev/mono/pkg/kube"
)

func (k *k8sSecretManager) getClient(ctx context.Context) (*kubernetes.Clientset, error) {
	var (
		kubeCfg *rest.Config
		err     error
	)

	if k.ClusterInfo != nil {
		kubeCfg, err = kube.ConfigForCluster(ctx, k.ClusterInfo)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get kube config")
		}
	} else {
		kubeCfg, err = kube.GetKubeConfig()
		if err != nil {
			return nil, fmt.Errorf("unable to get default kube config: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create kube client: %w", err)
	}

	return clientset, nil
}
