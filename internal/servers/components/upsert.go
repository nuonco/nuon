package components

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/powertoolsdev/api/internal/models"
	"github.com/powertoolsdev/api/internal/servers/converters"
	componentv1 "github.com/powertoolsdev/protos/api/generated/types/component/v1"
)

func (s *server) UpsertComponent(
	ctx context.Context,
	req *connect.Request[componentv1.UpsertComponentRequest],
) (*connect.Response[componentv1.UpsertComponentResponse], error) {
	component, err := s.Svc.UpsertComponent(ctx, models.ComponentInput{
		AppID: req.Msg.AppId,
		ID:    converters.ToOptionalStr(req.Msg.Id),
		Name:  req.Msg.Name,

		// NOTE: the following parameters will not be used once we migrate to the new component ref
		BuildImage: req.Msg.BuildImage,
		Type:       models.ComponentTypePublicImage,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to upsert component: %w", err)
	}

	return connect.NewResponse(&componentv1.UpsertComponentResponse{
		Component: converters.ComponentModelToProto(component),
	}), nil
}
