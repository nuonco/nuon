package status

import (
	"context"

	"github.com/bufbuild/connect-go"
	statusv1 "github.com/powertoolsdev/protos/orgs-api/generated/types/status/v1"
)

func (s *server) Ping(
	ctx context.Context,
	req *connect.Request[statusv1.PingRequest],
) (*connect.Response[statusv1.PingResponse], error) {
	res := connect.NewResponse(&statusv1.PingResponse{
		Status: "ok",
	})
	res.Header().Set("test-version", "v1")
	return res, nil
}
