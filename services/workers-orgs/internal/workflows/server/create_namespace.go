package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/kube"
	"go.temporal.io/sdk/activity"
	corev1 "k8s.io/api/core/v1"
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreapplyv1 "k8s.io/client-go/applyconfigurations/core/v1"
	"k8s.io/client-go/kubernetes"
)

var errOops error = fmt.Errorf("oops")

type CreateNamespaceRequest struct {
	NamespaceName string           `json:"namespace_name" validate:"required"`
	ClusterInfo   kube.ClusterInfo `json:"cluster_info" validate:"required"`
}

func (c CreateNamespaceRequest) validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

type CreateNamespaceResponse struct {
	NamespaceName string
}

type createNamespaceAPI interface {
	// Applies the desired namespace. Idempotent
	Apply(context.Context, *coreapplyv1.NamespaceApplyConfiguration, apimetav1.ApplyOptions) (*corev1.Namespace, error)
}

type namespaceCreator interface {
	createNamespace(context.Context, createNamespaceAPI, string) (*corev1.Namespace, error)
}

func (a *Activities) CreateNamespace(ctx context.Context, req CreateNamespaceRequest) (CreateNamespaceResponse, error) {
	resp := CreateNamespaceResponse{}
	l := activity.GetLogger(ctx)

	if err := req.validate(); err != nil {
		return resp, fmt.Errorf("invalid request: %w", err)
	}

	kCfg, err := a.getKubeConfig(&req.ClusterInfo)
	if err != nil {
		return resp, fmt.Errorf("unable to get kube config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(kCfg)
	if err != nil {
		return resp, fmt.Errorf("failed to create kube client: %w", err)
	}

	ns, err := a.createNamespace(ctx, clientset.CoreV1().Namespaces(), req.NamespaceName)
	if err != nil {
		return resp, fmt.Errorf("failed to create namespace: %w", err)
	}

	resp.NamespaceName = ns.Name
	l.Debug("finished creating namespace", "response", resp)
	return resp, nil
}

var (
	ErrInvalidNamespaceName = errors.New("invalid namespace name")
)

type nsCreator struct{}

var _ namespaceCreator = (*nsCreator)(nil)

func (n *nsCreator) createNamespace(ctx context.Context, api createNamespaceAPI, name string) (*corev1.Namespace, error) {
	nsOptions := coreapplyv1.Namespace(name)

	ns, err := api.Apply(ctx, nsOptions, apimetav1.ApplyOptions{
		FieldManager: "nuon-create-namespace-activity",
		Force:        true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to apply namespace ns: %s: %w", name, err)
	}

	return ns, nil
}
