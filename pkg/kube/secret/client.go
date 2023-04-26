package secret

import (
	"fmt"

	"github.com/powertoolsdev/mono/pkg/kube"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func (k *k8sSecretGetter) getClient() (kubeClientSecretGetter, error) {
	if k.client != nil {
		return k.client, nil
	}

	var (
		kubeCfg *rest.Config
		err     error
	)
	if k.ClusterInfo != nil {
		kubeCfg, err = kube.ConfigForCluster(k.ClusterInfo)
		if err != nil {
			return nil, fmt.Errorf("unable to get cluster kube config: %w", err)
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

	return clientset.CoreV1().Secrets(k.Namespace), nil
}
