package apps

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/powertoolsdev/mono/services/api/internal/servers/converters"
	appv1 "github.com/powertoolsdev/mono/pkg/protos/api/generated/types/app/v1"
)

func (s *server) GetApp(
	ctx context.Context,
	req *connect.Request[appv1.GetAppRequest],
) (*connect.Response[appv1.GetAppResponse], error) {
	// run protobuf validations
	if err := req.Msg.Validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	app, err := s.Svc.GetApp(ctx, req.Msg.Id)
	if err != nil {
		return nil, fmt.Errorf("unable to get app: %w", err)
	}

	return connect.NewResponse(&appv1.GetAppResponse{
		App: converters.AppModelToProto(app),
	}), nil
}

func (s *server) GetAppsByOrg(
	ctx context.Context,
	req *connect.Request[appv1.GetAppsByOrgRequest],
) (*connect.Response[appv1.GetAppsByOrgResponse], error) {
	// run protobuf validations
	if err := req.Msg.Validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	apps, _, err := s.Svc.GetOrgApps(ctx, req.Msg.OrgId, &models.ConnectionOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get apps: %w", err)
	}

	return connect.NewResponse(&appv1.GetAppsByOrgResponse{
		Apps: converters.AppModelsToProtos(apps),
	}), nil
}
