package installs

import (
	"context"

	"github.com/bufbuild/connect-go"
	installv1 "github.com/powertoolsdev/protos/api/generated/types/install/v1"
)

func (s *server) UpsertInstall(
	ctx context.Context,
	req *connect.Request[installv1.UpsertInstallRequest],
) (*connect.Response[installv1.UpsertInstallResponse], error) {
	res := connect.NewResponse(&installv1.UpsertInstallResponse{})
	return res, nil
}
