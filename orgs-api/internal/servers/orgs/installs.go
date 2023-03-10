package orgs

import (
	"context"

	"github.com/bufbuild/connect-go"
	orgsv1 "github.com/powertoolsdev/protos/orgs-api/generated/types/orgs/v1"
)

func (s *server) GetInstalls(
	ctx context.Context,
	req *connect.Request[orgsv1.GetInstallsRequest],
) (*connect.Response[orgsv1.GetInstallsResponse], error) {
	res := connect.NewResponse(&orgsv1.GetInstallsResponse{
		//
	})
	return res, nil
}
