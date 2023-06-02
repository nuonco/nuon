package apps

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	appv1 "github.com/powertoolsdev/mono/pkg/types/api/app/v1"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/powertoolsdev/mono/services/api/internal/servers/converters"
)

func (s *server) UpsertApp(
	ctx context.Context,
	req *connect.Request[appv1.UpsertAppRequest],
) (*connect.Response[appv1.UpsertAppResponse], error) {
	// run protobuf validations
	if err := req.Msg.Validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	app, err := s.Svc.UpsertApp(ctx, models.AppInput{
		ID:          converters.ToOptionalStr(req.Msg.Id),
		Name:        req.Msg.Name,
		OrgID:       req.Msg.OrgId,
		CreatedByID: &req.Msg.CreatedById,
		OverrideID:  &req.Msg.OverrideId,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to upsert app: %w", err)
	}

	return connect.NewResponse(&appv1.UpsertAppResponse{
		App: converters.AppModelToProto(app),
	}), nil
}
