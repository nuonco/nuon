package apps

import (
	"context"

	"github.com/bufbuild/connect-go"
	appv1 "github.com/powertoolsdev/protos/api/generated/types/app/v1"
)

func (s *server) DeleteApp(
	ctx context.Context,
	req *connect.Request[appv1.DeleteAppRequest],
) (*connect.Response[appv1.DeleteAppResponse], error) {
	res := connect.NewResponse(&appv1.DeleteAppResponse{})
	return res, nil
}
