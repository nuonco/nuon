package teardown

import (
	"context"
	"errors"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/kube"
	"go.temporal.io/sdk/activity"
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type DestroyNamespaceRequest struct {
	NamespaceName string
}

type DestroyNamespaceResponse struct{}

type destroyNamespaceAPI interface {
	Delete(context.Context, string, apimetav1.DeleteOptions) error
}

type namespaceDestroyer interface {
	destroyNamespace(context.Context, destroyNamespaceAPI, string) error
}

func (a *Activities) DestroyNamespace(ctx context.Context, req DestroyNamespaceRequest) (DestroyNamespaceResponse, error) {
	resp := DestroyNamespaceResponse{}
	l := activity.GetLogger(ctx)

	if err := validateDestroyNamespaceRequest(req); err != nil {
		return resp, fmt.Errorf("invalid request: %w", err)
	}

	var err error
	cfg := a.Kubeconfig
	if cfg == nil {
		cfg, err = kube.GetKubeConfig()
		if err != nil {
			return resp, fmt.Errorf("failed to get kube config: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return resp, fmt.Errorf("failed to create kube client: %w", err)
	}

	err = a.destroyNamespace(ctx, clientset.CoreV1().Namespaces(), req.NamespaceName)
	if err != nil {
		return resp, fmt.Errorf("failed to destroy namespace: %w", err)
	}

	l.Debug("finished destroying namespace", "response", resp)
	return resp, nil
}

var (
	ErrInvalidNamespaceName = errors.New("invalid namespace name")
)

func validateDestroyNamespaceRequest(req DestroyNamespaceRequest) error {
	if req.NamespaceName == "" {
		return fmt.Errorf("%w: kubernetes namespace must be specified", ErrInvalidNamespaceName)
	}

	return nil
}

type nsDestroyer struct{}

var _ namespaceDestroyer = (*nsDestroyer)(nil)

func (n *nsDestroyer) destroyNamespace(ctx context.Context, api destroyNamespaceAPI, name string) error {
	err := api.Delete(ctx, name, apimetav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete namespace ns: %s: %w", name, err)
	}

	return nil
}
