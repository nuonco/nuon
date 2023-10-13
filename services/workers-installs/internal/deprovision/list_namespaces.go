package deprovision

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/kube"
	"go.temporal.io/sdk/activity"
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

var terraformManagedNamespaces []string = []string{
	"alb-ingress",
	"cert-manager",
	"ebs-csi-controller",
	"external-dns",
	"kube-node-lease",
	"kube-public",
	"kube-system",
	"metrics-server",
	"nginx-ingress",
	"default",
}

type ListNamespacesRequest struct {
	OrgID     string
	AppID     string
	InstallID string
}

func (r ListNamespacesRequest) validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

type ListNamespacesResponse struct {
	Namespaces []string
}

func (a *Activities) ListNamespaces(ctx context.Context, req ListNamespacesRequest) (ListNamespacesResponse, error) {
	resp := ListNamespacesResponse{}
	l := activity.GetLogger(ctx)

	tfOutputs, err := a.getSandboxOutputs(ctx, req.OrgID, req.AppID, req.InstallID)
	if err != nil {
		return resp, fmt.Errorf("unable to get sandbox outputs: %w", err)
	}

	kubeCfg, err := a.getKubeConfig(&kube.ClusterInfo{
		ID:             tfOutputs.Cluster.Name,
		Endpoint:       tfOutputs.Cluster.Endpoint,
		CAData:         tfOutputs.Cluster.CertificateAuthorityData,
		TrustedRoleARN: a.cfg.NuonAccessRoleArn,
	})
	if err != nil {
		return resp, fmt.Errorf("unable to get kube config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		return resp, fmt.Errorf("failed to create kube client: %w", err)
	}

	namespaces, err := a.listNamespaces(ctx, clientset.CoreV1().Namespaces())
	if err != nil {
		return resp, fmt.Errorf("failed to destroy namespace: %w", err)
	}
	resp.Namespaces = namespaces

	l.Debug("finished destroying namespace", "response", resp)
	return resp, nil
}

func (a *Activities) listNamespaces(ctx context.Context, api corev1.NamespaceInterface) ([]string, error) {
	resp, err := api.List(ctx, apimetav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	namespaces := make([]string, 0)
	for _, namespace := range resp.Items {
		namespaces = append(namespaces, namespace.Name)
	}

	return namespaces, nil
}
