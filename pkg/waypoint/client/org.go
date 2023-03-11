package client

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	pb "github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/powertoolsdev/mono/pkg/waypoint/token"
	"google.golang.org/grpc"
)

type orgProvider struct {
	Address         string `validate:"required"` // NOTE(jdt): we should probably validate address format here
	SecretNamespace string `validate:"required"`
	SecretName      string `validate:"required"`

	// internal state
	v            *validator.Validate
	clientGetter clientGetter
	tokenGetter  tokenGetter
	clientConn   *grpc.ClientConn
}

type orgProviderOption func(*orgProvider) error

func NewOrgProvider(v *validator.Validate, opts ...orgProviderOption) (*orgProvider, error) {
	p := &orgProvider{v: v, clientGetter: &defaultClientGetter{}}

	if v == nil {
		return nil, fmt.Errorf("error instantiating org waypoint provider: validator is nil")
	}

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

func WithOrgConfig(cfg Config) orgProviderOption {
	return func(op *orgProvider) error {
		op.Address = cfg.Address
		op.SecretNamespace = cfg.Token.Namespace
		op.SecretName = cfg.Token.Name
		return nil
	}
}

func (p *orgProvider) Close() error {
	if p.clientConn == nil {
		return nil
	}

	return p.clientConn.Close()
}

// GetClient returns a waypoint client with the previously configured address,
// fetching the token from k8s using the token information
func (p *orgProvider) GetClient(ctx context.Context) (pb.WaypointClient, error) {
	getter := p.tokenGetter
	if getter == nil {
		k, err := token.New(p.v, token.WithNamespace(p.SecretNamespace), token.WithName(p.SecretName))
		if err != nil {
			return nil, err
		}
		getter = k
	}

	token, err := getter.GetOrgToken(ctx)
	if err != nil {
		return nil, err
	}

	cc, wp, err := p.clientGetter.getClient(ctx, p.Address, token)
	p.clientConn = cc
	return wp, err
}

// tokenGetter provides a way to get tokens for a specific org
type tokenGetter interface {
	GetOrgToken(context.Context) (string, error)
}
