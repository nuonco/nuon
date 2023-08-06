package introspection

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	memcached "k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

func (s *svc) getHelmCfg(ctx context.Context, namespace string) (*action.Configuration, error) {
	l := zap.L()
	actionLogger := func(format string, v ...interface{}) { l.Debug(fmt.Sprintf(format, v...)) }

	kubeCfg, err := s.getKubeConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get kube config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		return nil, fmt.Errorf("unable to get kube config: %w", err)
	}

	rcg := &RestClientGetter{RestConfig: kubeCfg, Clientset: clientset}
	actionCfg := new(action.Configuration)
	err = actionCfg.Init(rcg, namespace, "secret", actionLogger)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize action config: %w", err)
	}

	return actionCfg, nil
}

// RestClientGetter is an interface that helm requires to interact with it. Eventually, this should live in
// `pkg/helm/client`, but for now, we just provide it here.
type RestClientGetter struct {
	RestConfig *rest.Config
	Clientset  kubernetes.Interface
}

var _ genericclioptions.RESTClientGetter = (*RestClientGetter)(nil)

// ToRESTConfig implemented interface method
func (k *RestClientGetter) ToRESTConfig() (*rest.Config, error) {
	return k.RestConfig, nil
}

// ToDiscoveryClient implemented interface method
func (k *RestClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	return memcached.NewMemCacheClient(k.Clientset.Discovery()), nil
}

// ToRESTMapper implemented interface method
func (k *RestClientGetter) ToRESTMapper() (meta.RESTMapper, error) {
	discoveryClient, err := k.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
	expander := restmapper.NewShortcutExpander(mapper, discoveryClient)
	return expander, nil
}

// ToRawKubeConfigLoader implemented interface method
func (k *RestClientGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	// Build our config and client
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
}
