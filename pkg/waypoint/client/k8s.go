package client

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	pb "github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/powertoolsdev/mono/pkg/kube"
	"github.com/powertoolsdev/mono/pkg/kube/secret"
	"google.golang.org/grpc"
)

type k8sProvider struct {
	Address         string `validate:"required"`
	SecretNamespace string `validate:"required"`
	SecretName      string `validate:"required"`
	SecretKey       string `validate:"required"`

	ClusterInfo *kube.ClusterInfo

	// internal state
	v            *validator.Validate
	clientGetter clientGetter
	clientConn   *grpc.ClientConn
}

type k8sProviderOption func(*k8sProvider) error

func NewK8sProvider(v *validator.Validate, opts ...k8sProviderOption) (*k8sProvider, error) {
	p := &k8sProvider{v: v, clientGetter: &defaultClientGetter{}}

	for _, opt := range opts {
		if err := opt(p); err != nil {
			return nil, err
		}
	}

	if err := p.v.Struct(p); err != nil {
		return nil, err
	}

	return p, nil
}

func WithConfig(cfg Config) k8sProviderOption {
	return func(p *k8sProvider) error {
		p.Address = cfg.Address
		p.SecretNamespace = cfg.Token.Namespace
		p.SecretName = cfg.Token.Name
		p.SecretKey = cfg.Token.Key
		p.ClusterInfo = cfg.ClusterInfo
		return nil
	}
}

func (p *k8sProvider) Close() error {
	if p.clientConn == nil {
		return nil
	}

	return p.clientConn.Close()
}

// GetClient returns a waypoint client with the previously configured address,
// fetching the token from k8s using the token information
func (p *k8sProvider) GetClient(ctx context.Context) (pb.WaypointClient, error) {
	secretGetter, err := secret.New(p.v,
		secret.WithNamespace(p.SecretNamespace),
		secret.WithCluster(p.ClusterInfo),
		secret.WithName(p.SecretName),
		secret.WithKey(p.SecretKey),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to get secret getter: %w", err)
	}

	token, err := secretGetter.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get token secret: %w", err)
	}

	cc, wp, err := p.clientGetter.getClient(ctx, p.Address, string(token))
	p.clientConn = cc
	return wp, err
}
