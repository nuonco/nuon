package installs

import (
	"context"

	"github.com/bufbuild/connect-go"
	installv1 "github.com/powertoolsdev/protos/api/generated/types/install/v1"
)

func (s *server) GetInstall(
	ctx context.Context,
	req *connect.Request[installv1.GetInstallRequest],
) (*connect.Response[installv1.GetInstallResponse], error) {
	res := connect.NewResponse(&installv1.GetInstallResponse{})
	return res, nil
}

func (s *server) GetInstallsByApp(
	ctx context.Context,
	req *connect.Request[installv1.GetInstallsByAppRequest],
) (*connect.Response[installv1.GetInstallsByAppResponse], error) {
	res := connect.NewResponse(&installv1.GetInstallsByAppResponse{})
	return res, nil
}
