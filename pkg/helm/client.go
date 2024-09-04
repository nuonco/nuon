package helm

import (
	"fmt"

	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/powertoolsdev/mono/pkg/deprecated/helm"
)

const (
	defaultHelmDriver string = "secret"
)

func Client(log *zap.Logger, kubeCfg *rest.Config, ns string) (*action.Configuration, error) {
	clientset, err := kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create kube client: %w", err)
	}

	debug := func(format string, vs ...interface{}) {
		msg := fmt.Sprintf(format, vs...)
		log.Info(msg)
	}

	// Initialize our action
	var ac action.Configuration
	err = ac.Init(&helm.RestClientGetter{
		RestConfig: kubeCfg,
		Clientset:  clientset,
	}, ns, defaultHelmDriver, debug)
	if err != nil {
		return nil, fmt.Errorf("unable to get rest client: %w", err)
	}

	return &ac, nil
}
