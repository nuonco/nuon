package status

import (
	"context"

	"github.com/bufbuild/connect-go"
	statusv1 "github.com/powertoolsdev/protos/api/generated/types/status/v1"
)

func (s *server) Version(
	ctx context.Context,
	req *connect.Request[statusv1.VersionRequest],
) (*connect.Response[statusv1.VersionResponse], error) {
	res := connect.NewResponse(&statusv1.VersionResponse{
		GitRef: s.GitRef,
	})
	return res, nil
}
