package kubernetes_manifest

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"

	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/pkg/kube"
)

type kubernetesClient struct {
	client          *dynamic.DynamicClient
	discoveryMapper *restmapper.DeferredDiscoveryRESTMapper
	discoveryClient discovery.ServerResourcesInterface
}

func (k *kubernetesClient) resourcesforGroupVersion(gv string) (*metav1.APIResourceList, error) {
	if len(gv) != 0 && gv[0] == '/' {
		gv = gv[1:] // remove leading slash if present
	}
	return k.discoveryClient.ServerResourcesForGroupVersion(gv)
}

func (h *handler) getClient(ctx context.Context) (*kubernetesClient, error) {
	kubeCfg, err := kube.ConfigForCluster(ctx, h.state.plan.KubernetesManifestDeployPlan.ClusterInfo)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get kube config")
	}

	dynamicClient, err := dynamic.NewForConfig(kubeCfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get kubernetes dynamic client")
	}

	clientSet, err := kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get kubernetes client set")
	}

	dmc := memory.NewMemCacheClient(clientSet.Discovery())
	discoveryMapper := restmapper.NewDeferredDiscoveryRESTMapper(dmc)

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(kubeCfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get discovery client")
	}

	return &kubernetesClient{
		client:          dynamicClient,
		discoveryMapper: discoveryMapper,
		discoveryClient: discoveryClient,
	}, nil
}
