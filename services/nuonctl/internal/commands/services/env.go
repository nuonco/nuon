package services

import (
	"context"
	"fmt"

	eksclient "github.com/powertoolsdev/mono/pkg/aws/eks-client"
	"github.com/powertoolsdev/mono/services/nuonctl/internal/json"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

//
//go:generate -command mockgen go run github.com/golang/mock/mockgen
//go:generate mockgen -destination=env_mock_test.go -source=env.go -package=services
const (
	region                = "us-west-2"
	clusterName           = "stage-nuon"
	assumeRoleSessionName = "nuonctl"
	assumeRoleARN         = ""
	serviceNamespace      = "default"
)

func (c *commands) Env(ctx context.Context, svc string) error {
	eksClienter, err := eksclient.New(c.v,
		eksclient.WithClusterName(clusterName),
		eksclient.WithRegion(region),
		eksclient.WithRoleSessionName(assumeRoleSessionName),
		eksclient.WithRoleARN(assumeRoleARN),
	)
	if err != nil {
		return fmt.Errorf("unable to get eks client creator: %w", err)
	}

	kubeCfg, err := eksClienter.GetKubeConfig(ctx)
	if err != nil {
		return fmt.Errorf("unable to get kube config: %w", err)
	}

	kubeClient, err := kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		return fmt.Errorf("failed to create kube client: %w", err)
	}

	configMapsClient := kubeClient.CoreV1().ConfigMaps(serviceNamespace)
	env, err := c.getServiceEnv(ctx, configMapsClient, svc)
	if err != nil {
		return fmt.Errorf("unable to get config map: %w", err)
	}

	return json.Print(env)
}

type k8sConfigMapGetter interface {
	Get(context.Context, string, metav1.GetOptions) (*corev1.ConfigMap, error)
}

func (c *commands) getServiceEnv(ctx context.Context, client k8sConfigMapGetter, name string) (map[string]string, error) {
	cfgMap, err := client.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get config map: %w", err)
	}

	return cfgMap.Data, nil
}
