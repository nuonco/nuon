package server

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/kube"
	"go.temporal.io/sdk/activity"
	corev1 "k8s.io/api/core/v1"
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	coreapplyv1 "k8s.io/client-go/applyconfigurations/core/v1"
	metaapplyv1 "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ExposeWaypointServerRequest struct {
	NamespaceName string           `json:"namespace_name" validate:"required"`
	RootDomain    string           `json:"root_domain" validate:"required"`
	ShortID       string           `json:"short_id" validate:"required"`
	ClusterInfo   kube.ClusterInfo `json:"cluster_info" validate:"required"`
}

func (e ExposeWaypointServerRequest) validate() error {
	validate := validator.New()
	return validate.Struct(e)
}

type ExposeWaypointServerResponse struct{}

func (a *Activities) ExposeWaypointServer(ctx context.Context, req ExposeWaypointServerRequest) (ExposeWaypointServerResponse, error) {
	resp := ExposeWaypointServerResponse{}
	l := activity.GetLogger(ctx)

	if err := req.validate(); err != nil {
		return resp, err
	}

	var err error
	kCfg := a.Kubeconfig
	if kCfg == nil {
		kCfg, err = kube.ConfigForCluster(&req.ClusterInfo)
		if err != nil {
			return resp, fmt.Errorf("failed to get config for cluster: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(kCfg)
	if err != nil {
		return resp, fmt.Errorf("failed to create kube client: %w", err)
	}

	svc, err := a.createService(ctx, clientset.CoreV1().Services(req.NamespaceName), req)
	if err != nil {
		return resp, fmt.Errorf("failed to create service: %w", err)
	}

	l.Debug("finished creating service", svc.Name)
	return resp, nil
}

// k8sServiceCreator is the interface to kubernetes that we use to actually create the service
type k8sServiceCreator interface {
	Apply(context.Context, *coreapplyv1.ServiceApplyConfiguration, apimetav1.ApplyOptions) (*corev1.Service, error)
}

type serviceCreator interface {
	createService(context.Context, k8sServiceCreator, ExposeWaypointServerRequest) (*corev1.Service, error)
}

type svcCreator struct{}

var _ serviceCreator = (*svcCreator)(nil)

func (s *svcCreator) createService(ctx context.Context, api k8sServiceCreator, req ExposeWaypointServerRequest) (*corev1.Service, error) {
	name := fmt.Sprintf("wp-%s-waypoint-server-public", req.ShortID)
	hostname := fmt.Sprintf("%s.%s", req.ShortID, req.RootDomain)
	cfg := coreapplyv1.Service(name, req.NamespaceName)

	cfg.ObjectMetaApplyConfiguration = &metaapplyv1.ObjectMetaApplyConfiguration{
		Name:      &name,
		Namespace: &req.NamespaceName,
		Annotations: map[string]string{
			"external-dns.alpha.kubernetes.io/hostname":                            hostname,
			"service.beta.kubernetes.io/aws-load-balancer-nlb-target-type":         "ip",
			"service.beta.kubernetes.io/aws-load-balancer-scheme":                  "internet-facing",
			"service.beta.kubernetes.io/aws-load-balancer-target-group-attributes": "preserve_client_ip.enabled=false",
		},
		Labels: map[string]string{
			"app.kubernetes.io/managed-by": "nuon",
		},
	}

	cfg.Spec = &coreapplyv1.ServiceSpecApplyConfiguration{
		Type:                          generics.ToPtr(corev1.ServiceTypeLoadBalancer),
		LoadBalancerClass:             generics.ToPtr("service.k8s.aws/nlb"),
		AllocateLoadBalancerNodePorts: generics.ToPtr(false),
		ExternalTrafficPolicy:         generics.ToPtr(corev1.ServiceExternalTrafficPolicyTypeLocal),
		InternalTrafficPolicy:         generics.ToPtr(corev1.ServiceInternalTrafficPolicyLocal),
		Selector: map[string]string{
			"app.kubernetes.io/instance": fmt.Sprintf("wp-%s", req.ShortID),
			"app.kubernetes.io/name":     "waypoint",
			"component":                  "server",
		},
		Ports: []coreapplyv1.ServicePortApplyConfiguration{
			{
				Name:       generics.ToPtr("grpc"),
				TargetPort: generics.ToPtr(intstr.FromString("grpc")),
				Port:       generics.ToPtr(defaultWaypointServerPort),
			},
		},
	}

	svc, err := api.Apply(ctx, cfg, apimetav1.ApplyOptions{
		FieldManager: "nuon-expose-waypoint-server-activity",
		Force:        true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to apply service svc: %s: %w", name, err)
	}

	return svc, nil
}
