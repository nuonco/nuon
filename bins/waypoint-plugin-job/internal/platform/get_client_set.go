package platform

import (
	"github.com/powertoolsdev/mono/pkg/kube"
	"k8s.io/client-go/kubernetes"
)

func (p *Platform) getClientset() (*kubernetes.Clientset, error) {
	config, err := kube.GetKubeConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}
