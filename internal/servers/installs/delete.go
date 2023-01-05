package installs

import (
	"context"

	"github.com/bufbuild/connect-go"
	installv1 "github.com/powertoolsdev/protos/api/generated/types/install/v1"
)

func (s *server) DeleteInstall(
	ctx context.Context,
	req *connect.Request[installv1.DeleteInstallRequest],
) (*connect.Response[installv1.DeleteInstallResponse], error) {
	res := connect.NewResponse(&installv1.DeleteInstallResponse{})
	return res, nil
}
