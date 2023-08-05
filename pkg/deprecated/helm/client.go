package helm

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	memcached "k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	helmDriver = "secret"
)

type Logger interface {
	Debug(msg string, keyvals ...interface{})
}

type fmtLogger struct{}

func (f fmtLogger) Debug(msg string, keyvals ...interface{}) {
	fmt.Println(msg, keyvals)
}

// Helm has this silly interface for k8s clients. It's painful to use in-cluster.
// This is lightly modified from https://github.com/hashicorp/waypoint/blob/main/builtin/k8s/helm/rest.go

// RestClientGetter is a RESTClientGetter interface implementation for the
// Helm Go packages.
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
