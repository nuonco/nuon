package components

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/powertoolsdev/api/internal/models"
	"github.com/powertoolsdev/api/internal/servers/converters"
	componentv1 "github.com/powertoolsdev/protos/api/generated/types/component/v1"
)

func (s *server) GetComponent(
	ctx context.Context,
	req *connect.Request[componentv1.GetComponentRequest],
) (*connect.Response[componentv1.GetComponentResponse], error) {
	// run protobuf validations
	if err := req.Msg.Validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	component, err := s.Svc.GetComponent(ctx, req.Msg.Id)
	if err != nil {
		return nil, fmt.Errorf("unable to get component: %w", err)
	}

	return connect.NewResponse(&componentv1.GetComponentResponse{
		Component: converters.ComponentModelToProto(component),
	}), nil
}

func (s *server) GetComponentsByApp(
	ctx context.Context,
	req *connect.Request[componentv1.GetComponentsByAppRequest],
) (*connect.Response[componentv1.GetComponentsByAppResponse], error) {
	// run protobuf validations
	if err := req.Msg.Validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	components, _, err := s.Svc.GetAppComponents(ctx, req.Msg.AppId, &models.ConnectionOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get app components: %w", err)
	}

	return connect.NewResponse(&componentv1.GetComponentsByAppResponse{
		Components: converters.ComponentModelsToProtos(components),
	}), nil
}
