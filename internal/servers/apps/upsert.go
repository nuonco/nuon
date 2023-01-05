package apps

import (
	"context"

	"github.com/bufbuild/connect-go"
	appv1 "github.com/powertoolsdev/protos/api/generated/types/app/v1"
)

func (s *server) UpsertApp(
	ctx context.Context,
	req *connect.Request[appv1.UpsertAppRequest],
) (*connect.Response[appv1.UpsertAppResponse], error) {
	res := connect.NewResponse(&appv1.UpsertAppResponse{})
	return res, nil
}
