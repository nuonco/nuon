package helm

import (
	"fmt"

	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/kube"
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

	debug := func(format string, vs ...interface{}) {
		msg := fmt.Sprintf(format, vs...)
		log.Info(msg)
	}

	// Initialize our action
	var ac action.Configuration
	err = ac.Init(&RestClientGetter{
		RestConfig: kubeCfg,
		Clientset:  clientset,
		Namespace:  ns,
	}, ns, defaultHelmDriver, debug)
	if err != nil {
		return nil, fmt.Errorf("unable to get rest client: %w", err)
	}

	return &ac, nil
}

// ClientV2 initializes a new Helm client with the given logger and kube config.
// NOTE: it doesn't initialise the release store.
func ClientV2(log *zap.Logger, kubeCfg *rest.Config) (*action.Configuration, error) {
	clientset, err := kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create kube client: %w", err)
	}

	debug := func(format string, vs ...interface{}) {
		msg := fmt.Sprintf(format, vs...)
		log.Info(msg)
	}

	// Initialize our action
	ac, err := initActionConfig(&RestClientGetter{
		RestConfig: kubeCfg,
		Clientset:  clientset,
	}, debug)

	if err != nil {
		return nil, fmt.Errorf("unable to get rest client: %w", err)
	}

	return ac, nil
}

func initActionConfig(getter *RestClientGetter, log action.DebugLog) (*action.Configuration, error) {
	actionCfg := action.Configuration{}

	kc := kube.New(getter)

	actionCfg.RESTClientGetter = getter
	actionCfg.KubeClient = kc
	actionCfg.Log = log

	return &actionCfg, nil
}
