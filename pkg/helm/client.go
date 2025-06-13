package helm

import (
	"fmt"
	"log/slog"

	"go.uber.org/zap"
	"helm.sh/helm/v4/pkg/action"
	"helm.sh/helm/v4/pkg/kube"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	defaultHelmDriver string = "secret"
)

func Client(log *zap.Logger, kubeCfg *rest.Config, ns string) (*action.Configuration, error) {
	clientset, err := kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create kube client: %w", err)
	}
	logger := NewLogger(func() bool {
		return true
	})
	slog.SetDefault(logger)
	// Initialize our action
	var ac action.Configuration
	err = ac.Init(&RestClientGetter{
		RestConfig: kubeCfg,
		Clientset:  clientset,
		Namespace:  ns,
	}, ns, defaultHelmDriver)
	if err != nil {
		return nil, fmt.Errorf("unable to get rest client: %w", err)
	}

	return &ac, nil
}

// ClientV2 initializes a new Helm client with the given logger and kube config.
// NOTE: it doesn't initialise the release store.
func ClientV2(log *zap.Logger, kubeCfg *rest.Config, ns string) (*action.Configuration, error) {
	clientset, err := kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create kube client: %w", err)
	}

	// Initialize our action
	ac, err := initActionConfig(&RestClientGetter{
		RestConfig: kubeCfg,
		Clientset:  clientset,
		Namespace:  ns,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to get rest client: %w", err)
	}

	return ac, nil
}

func initActionConfig(getter *RestClientGetter) (*action.Configuration, error) {
	actionCfg := action.Configuration{}

	kc := kube.New(getter)

	actionCfg.RESTClientGetter = getter
	actionCfg.KubeClient = kc

	return &actionCfg, nil
}
