package deprovision

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/kube"
	"go.temporal.io/sdk/activity"
	"k8s.io/apimachinery/pkg/api/errors"
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type DestroyNamespaceRequest struct {
	NamespaceName string           `validate:"required"`
	ClusterInfo   kube.ClusterInfo `validate:"required"`
}

func (r DestroyNamespaceRequest) validate() error {
	validate := validator.New()
	return validate.Struct(r)
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

	kubeCfg, err := a.getKubeConfig(&req.ClusterInfo)
	if err != nil {
		return resp, fmt.Errorf("unable to get kube config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		return resp, fmt.Errorf("failed to create kube client: %w", err)
	}

	err = a.destroyNamespace(ctx, clientset.CoreV1().Namespaces(), req.NamespaceName)
	if err != nil {
		if errors.IsNotFound(err) {
			return resp, nil
		}

		return resp, fmt.Errorf("failed to destroy namespace: %w", err)
	}

	l.Debug("finished destroying namespace", "response", resp)
	return resp, nil
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
