package job

import (
	"k8s.io/client-go/kubernetes"

	"github.com/powertoolsdev/mono/pkg/kube"
)

func (p *handler) getClientset() (*kubernetes.Clientset, error) {
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
