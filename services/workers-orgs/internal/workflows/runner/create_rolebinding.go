package runner

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/kube"
	rbacv1 "k8s.io/api/rbac/v1"
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rbacapplyv1 "k8s.io/client-go/applyconfigurations/rbac/v1"
	"k8s.io/client-go/kubernetes"
)

type CreateRoleBindingRequest struct {
	TokenSecretNamespace string `json:"token_secret_namespace" validate:"required"`
	OrgServerAddr        string `json:"org_server_address" validate:"required"`
	OrgID                string `json:"org_id" validate:"required"`
	NamespaceName        string `json:"namespace_name" validate:"required"`

	ClusterInfo kube.ClusterInfo `json:"cluster_info" validate:"required"`
}

func (c CreateRoleBindingRequest) validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

type CreateRoleBindingResponse struct{}

var _ roleBindingCreator = (*roleBindingCreatorImpl)(nil)

func (a *Activities) CreateRoleBinding(
	ctx context.Context,
	req CreateRoleBindingRequest,
) (CreateRoleBindingResponse, error) {
	var resp CreateRoleBindingResponse
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

	if err := a.createRoleBinding(ctx, clientset.RbacV1().RoleBindings(req.NamespaceName), req); err != nil {
		return resp, fmt.Errorf("failed to create role binding: %w", err)
	}

	return resp, nil
}

// k8sRoleBindingCreator is the interface to kubernetes that we use to actually create the service
type k8sRoleBindingCreator interface {
	Apply(context.Context, *rbacapplyv1.RoleBindingApplyConfiguration, apimetav1.ApplyOptions) (*rbacv1.RoleBinding, error)
}

type roleBindingCreator interface {
	createRoleBinding(context.Context, k8sRoleBindingCreator, CreateRoleBindingRequest) error
}

type roleBindingCreatorImpl struct{}

func (roleBindingCreatorImpl) createRoleBinding(ctx context.Context, client k8sRoleBindingCreator, req CreateRoleBindingRequest) error {
	name := fmt.Sprintf("wp-%s-waypoint-runner-rolebinding", req.OrgID)
	svcAccountName := runnerServiceAccountName(req.OrgID)

	rb := rbacapplyv1.RoleBinding(name, req.NamespaceName)
	rb.WithLabels(map[string]string{
		"managed-by": "nuon",
	})
	rb.Subjects = []rbacapplyv1.SubjectApplyConfiguration{
		{
			Kind:      generics.ToPtr("ServiceAccount"),
			Name:      &svcAccountName,
			Namespace: &req.NamespaceName,
		},
	}
	rb.RoleRef = &rbacapplyv1.RoleRefApplyConfiguration{
		APIGroup: generics.ToPtr("rbac.authorization.k8s.io"),
		Kind:     generics.ToPtr("ClusterRole"),
		Name:     generics.ToPtr("edit"),
	}

	_, err := client.Apply(ctx, rb, apimetav1.ApplyOptions{
		FieldManager: "nuon-create-role-binding-activity",
		Force:        true,
	})
	if err != nil {
		return fmt.Errorf("failed to create role binding: %s: %w", name, err)
	}

	return nil
}
