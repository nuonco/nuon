package apps

import (
	"context"

	"github.com/bufbuild/connect-go"
	appv1 "github.com/powertoolsdev/protos/api/generated/types/app/v1"
)

func (s *server) GetApp(
	ctx context.Context,
	req *connect.Request[appv1.GetAppRequest],
) (*connect.Response[appv1.GetAppResponse], error) {
	res := connect.NewResponse(&appv1.GetAppResponse{})
	return res, nil
}

func (s *server) GetAppsByOrg(
	ctx context.Context,
	req *connect.Request[appv1.GetAppsByOrgRequest],
) (*connect.Response[appv1.GetAppsByOrgResponse], error) {
	res := connect.NewResponse(&appv1.GetAppsByOrgResponse{})
	return res, nil
}
