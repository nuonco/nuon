package public

import (
	"context"

	"github.com/go-playground/validator/v10"
	pb "github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/powertoolsdev/mono/pkg/waypoint/client"
	"google.golang.org/grpc"
)

var _ client.Provider = (*publicClient)(nil)

type publicClient struct {
	Address string `validate:"required"` // NOTE(jdt): we should probably validate address format here

	v *validator.Validate

	// internal state
	clientConn *grpc.ClientConn
}

type publicClientOption func(*publicClient) error

func New(v *validator.Validate, opts ...publicClientOption) (*publicClient, error) {
	c := &publicClient{v: v}
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

func WithAddress(addr string) publicClientOption {
	return func(up *publicClient) error {
		up.Address = addr
		return nil
	}
}

func (c *publicClient) Close() error {
	if c.clientConn == nil {
		return nil
	}

	return c.clientConn.Close()
}

// Fetch returns a waypoint client with no configured token
func (c *publicClient) Fetch(ctx context.Context) (pb.WaypointClient, error) {
	cc, err := c.getClient(ctx, c.Address)
	c.clientConn = cc
	client := pb.NewWaypointClient(cc)
	return client, err
}
