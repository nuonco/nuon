package k8s

import (
	"io/ioutil"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type clientsetInfo struct {
	Clientset *kubernetes.Clientset
	Namespace string
	Config    *rest.Config
}

func Clientset(kubeconfig, context string) (*kubernetes.Clientset, string, *rest.Config, error) {
	if kubeconfig == "" {
		cs, ns, c, err := ClientsetInCluster()
		if err == nil {
			return cs, ns, c, nil
		}

		if err != rest.ErrNotInCluster {
			return nil, "", nil, err
		}
	}

	return ClientsetOutOfCluster(kubeconfig, context)
}

func ClientsetOutOfCluster(kubeconfig, context string) (*kubernetes.Clientset, string, *rest.Config, error) {
	loader := clientcmd.NewDefaultClientConfigLoadingRules()

	if kubeconfig != "" {
		loader.ExplicitPath = kubeconfig
	}

	config := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loader,
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		},
	)

	ns, _, err := config.Namespace()
	if err != nil {
		return nil, "", nil, status.Errorf(codes.Aborted,
			"failed to initialize K8S client configuration: %s", err)
	}

	clientconfig, err := config.ClientConfig()
	if err != nil {
		return nil, "", nil, status.Errorf(codes.Aborted,
			"failed to initialize K8S client configuration: %s", err)
	}

	clientset, err := kubernetes.NewForConfig(clientconfig)
	if err != nil {
		return nil, "", nil, status.Errorf(codes.Aborted,
			"failed to initialize K8S client: %s", err)
	}

	return clientset, ns, clientconfig, nil
}

func ClientsetInCluster() (*kubernetes.Clientset, string, *rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, "", nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, "", nil, err
	}

	ns := "default"
	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if v := strings.TrimSpace(string(data)); len(v) > 0 {
			ns = v
		}
	}

	return clientset, ns, config, nil
}
