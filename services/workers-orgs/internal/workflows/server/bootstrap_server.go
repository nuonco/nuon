package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/gogo/status"
	pb "github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/kube"
	"github.com/powertoolsdev/mono/pkg/waypoint/client/public"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"
	"k8s.io/client-go/kubernetes"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	coreapplyv1 "k8s.io/client-go/applyconfigurations/core/v1"
	metaapplyv1 "k8s.io/client-go/applyconfigurations/meta/v1"
)

type BootstrapWaypointServerRequest struct {
	ServerAddr     string `json:"server_addr"     validate:"required"`
	TokenNamespace string `json:"token_namespace" validate:"required"`
	OrgID          string `json:"org_id" validate:"required"`

	ClusterInfo kube.ClusterInfo `json:"cluster_info" validate:"required"`
}

func (b BootstrapWaypointServerRequest) validate() error {
	validate := validator.New()
	return validate.Struct(b)
}

type BootstrapWaypointServerResponse struct{}

// BootstrapWaypointServer calls the bootstrap method on a server, grabs the bootstrap token and stores it in a
// kubernetes secret
func (a *Activities) BootstrapWaypointServer(
	ctx context.Context,
	req BootstrapWaypointServerRequest,
) (BootstrapWaypointServerResponse, error) {
	var resp BootstrapWaypointServerResponse

	provider, err := public.New(a.v, public.WithAddress(req.ServerAddr))
	if err != nil {
		return resp, fmt.Errorf("unable to get waypoint provider: %w", err)
	}

	client, err := provider.Fetch(ctx)
	if err != nil {
		return resp, fmt.Errorf("unable to get client: %w", err)
	}

	bootstrapToken, err := a.bootstrapWaypointServer(ctx, client)
	if errors.Is(err, BootstrapError{}) {
		return resp, nil
	} else if err != nil {
		return resp, err
	}

	cfg, err := kube.ConfigForCluster(&req.ClusterInfo)
	if err != nil {
		return resp, fmt.Errorf("failed to get config for cluster: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return resp, fmt.Errorf("failed to create kube client: %w", err)
	}

	err = a.storeBootstrapToken(ctx, clientset.CoreV1().Secrets(req.TokenNamespace), req.OrgID, bootstrapToken)
	if err != nil {
		return resp, fmt.Errorf("failed to create namespace: %w", err)
	}

	return resp, nil
}

type waypointServerBootstrapper interface {
	bootstrapWaypointServer(context.Context, waypointClientBootstrapper) (string, error)
	storeBootstrapToken(context.Context, kubeClientSecretStorer, string, string) error
}

var _ waypointServerBootstrapper = (*wpServerBootstrapper)(nil)

type wpServerBootstrapper struct{}

type waypointClientBootstrapper interface {
	BootstrapToken(context.Context, *emptypb.Empty, ...grpc.CallOption) (*pb.NewTokenResponse, error)
}

func (w *wpServerBootstrapper) bootstrapWaypointServer(ctx context.Context, client waypointClientBootstrapper) (string, error) {
	resp, err := client.BootstrapToken(ctx, &emptypb.Empty{})
	statusCode, ok := status.FromError(err)
	if ok && statusCode.Code() == codes.PermissionDenied {
		return "", BootstrapError{}
	} else if err != nil {
		return "", err
	}

	return resp.Token, nil
}

type kubeClientSecretStorer interface {
	Apply(context.Context, *coreapplyv1.SecretApplyConfiguration, metav1.ApplyOptions) (*corev1.Secret, error)
}

func getTokenSecretName(orgID string) string {
	return fmt.Sprintf("waypoint-bootstrap-token-%v", orgID)
}

func (w *wpServerBootstrapper) storeBootstrapToken(ctx context.Context, client kubeClientSecretStorer, id, token string) error {
	secretName := getTokenSecretName(id)
	secret := &coreapplyv1.SecretApplyConfiguration{
		TypeMetaApplyConfiguration: metaapplyv1.TypeMetaApplyConfiguration{
			Kind:       generics.ToPtr("Secret"),
			APIVersion: generics.ToPtr("v1"),
		},
		ObjectMetaApplyConfiguration: &metaapplyv1.ObjectMetaApplyConfiguration{
			Name: generics.ToPtr(secretName),
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "nuon",
			},
		},
		StringData: map[string]string{
			"token": token,
		},
		Type: generics.ToPtr(corev1.SecretTypeOpaque),
	}
	applyOpts := metav1.ApplyOptions{
		FieldManager: "nuon-store-bootstrap-token-activity",
		Force:        true,
	}

	_, err := client.Apply(ctx, secret, applyOpts)
	if err != nil {
		return fmt.Errorf("failed to store secret %v: %w", secretName, err)
	}
	return nil
}
