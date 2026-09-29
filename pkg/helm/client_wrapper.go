package helm

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	memcached "k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

type RestClientGetter struct {
	Namespace  string
	RestConfig *rest.Config
	Clientset  kubernetes.Interface
}

var _ genericclioptions.RESTClientGetter = (*RestClientGetter)(nil)

func (k *RestClientGetter) ToRESTConfig() (*rest.Config, error) {
	return k.RestConfig, nil
}

func (k *RestClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	return memcached.NewMemCacheClient(k.Clientset.Discovery()), nil
}

func (k *RestClientGetter) ToRESTMapper() (meta.RESTMapper, error) {
	discoveryClient, err := k.ToDiscoveryClient()
	if err != nil {
		return nil, err
	}

	mapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
	expander := restmapper.NewShortcutExpander(mapper, discoveryClient, func(string) {
	})
	return expander, nil
}

func (k *RestClientGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{
			Context: clientcmdapi.Context{
				Namespace: k.Namespace,
			},
		},
	)
}
