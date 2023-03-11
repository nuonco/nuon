package components

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/powertoolsdev/mono/services/api/internal/servers/converters"
	componentv1 "github.com/powertoolsdev/mono/pkg/protos/api/generated/types/component/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

func (s *server) UpsertComponent(
	ctx context.Context,
	req *connect.Request[componentv1.UpsertComponentRequest],
) (*connect.Response[componentv1.UpsertComponentResponse], error) {
	// run protobuf validations
	//if err := req.Msg.Validate(); err != nil {
	//return nil, fmt.Errorf("input validation failed: %w", err)
	//}

	params := models.ComponentInput{
		AppID:       req.Msg.AppId,
		ID:          converters.ToOptionalStr(req.Msg.Id),
		Name:        req.Msg.Name,
		CreatedByID: req.Msg.CreatedById,
	}

	// convert input ComponentConfig to JSON
	if req.Msg.ComponentConfig != nil {
		componentConfig, err := protojson.Marshal(req.Msg.ComponentConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to parse component configuration: %w", err)
		}
		params.Config = componentConfig
	}

	component, err := s.Svc.UpsertComponent(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert component: %w", err)
	}

	componentToProto, err := converters.ComponentModelToProto(component)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to proto: %w", err)
	}

	return connect.NewResponse(&componentv1.UpsertComponentResponse{
		Component: componentToProto,
	}), nil
}
