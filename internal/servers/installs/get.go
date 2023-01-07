package installs

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/powertoolsdev/api/internal/models"
	"github.com/powertoolsdev/api/internal/servers/converters"
	installv1 "github.com/powertoolsdev/protos/api/generated/types/install/v1"
)

func (s *server) GetInstall(
	ctx context.Context,
	req *connect.Request[installv1.GetInstallRequest],
) (*connect.Response[installv1.GetInstallResponse], error) {
	install, err := s.Svc.GetInstall(ctx, req.Msg.Id)
	if err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	return connect.NewResponse(&installv1.GetInstallResponse{
		Install: converters.InstallModelToProto(install),
	}), nil
}

func (s *server) GetInstallsByApp(
	ctx context.Context,
	req *connect.Request[installv1.GetInstallsByAppRequest],
) (*connect.Response[installv1.GetInstallsByAppResponse], error) {
	installs, _, err := s.Svc.GetAppInstalls(ctx, req.Msg.AppId, &models.ConnectionOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get installs: %w", err)
	}

	return connect.NewResponse(&installv1.GetInstallsByAppResponse{
		Installs: converters.InstallModelsToProtos(installs),
	}), nil
}
