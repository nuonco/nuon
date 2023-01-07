package apps

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/powertoolsdev/api/internal/models"
	"github.com/powertoolsdev/api/internal/servers/converters"
	appv1 "github.com/powertoolsdev/protos/api/generated/types/app/v1"
)

func (s *server) UpsertApp(
	ctx context.Context,
	req *connect.Request[appv1.UpsertAppRequest],
) (*connect.Response[appv1.UpsertAppResponse], error) {
	app, err := s.Svc.UpsertApp(ctx, models.AppInput{
		ID:              converters.ToOptionalStr(req.Msg.Id),
		Name:            req.Msg.Name,
		OrgID:           req.Msg.OrgId,
		GithubInstallID: converters.ToOptionalStr(req.Msg.GithubInstallId),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to upsert app: %w", err)
	}

	return connect.NewResponse(&appv1.UpsertAppResponse{
		App: converters.AppModelToProto(app),
	}), nil
}
