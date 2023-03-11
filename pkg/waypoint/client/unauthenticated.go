package client

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	pb "github.com/hashicorp/waypoint/pkg/server/gen"
	"google.golang.org/grpc"
)

type unauthedProvider struct {
	Address string `validate:"required"` // NOTE(jdt): we should probably validate address format here

	// internal state
	v            *validator.Validate
	clientGetter clientGetter
	clientConn   *grpc.ClientConn
}

type unauthedProviderOption func(*unauthedProvider) error

func NewUnauthenticatedProvider(v *validator.Validate, opts ...unauthedProviderOption) (*unauthedProvider, error) {
	c := &unauthedProvider{v: v, clientGetter: &defaultClientGetter{}}

	if v == nil {
		return nil, fmt.Errorf("error instantiating unauthenticated waypoint provider: validator is nil")
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	if err := c.v.Struct(c); err != nil {
		return nil, err
	}

	return c, nil
}

func WithUnauthenticatedConfig(cfg Config) unauthedProviderOption {
	return func(up *unauthedProvider) error {
		up.Address = cfg.Address
		return nil
	}
}

func (c *unauthedProvider) Close() error {
	if c.clientConn == nil {
		return nil
	}

	return c.clientConn.Close()
}

// GetUnauthenticatedWaypointClient returns a waypoint client with no configured token
func (c *unauthedProvider) GetClient(ctx context.Context) (pb.WaypointClient, error) {
	cc, wp, err := c.clientGetter.getClient(ctx, c.Address, "")
	c.clientConn = cc
	return wp, err
}
