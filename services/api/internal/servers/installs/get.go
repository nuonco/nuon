package installs

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	installv1 "github.com/powertoolsdev/mono/pkg/types/api/install/v1"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/powertoolsdev/mono/services/api/internal/servers/converters"
)

func (s *server) GetInstall(
	ctx context.Context,
	req *connect.Request[installv1.GetInstallRequest],
) (*connect.Response[installv1.GetInstallResponse], error) {
	// run protobuf validations
	if err := req.Msg.Validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

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
	// run protobuf validations
	if err := req.Msg.Validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	installs, _, err := s.Svc.GetAppInstalls(ctx, req.Msg.AppId, &models.ConnectionOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to get installs: %w", err)
	}

	return connect.NewResponse(&installv1.GetInstallsByAppResponse{
		Installs: converters.InstallModelsToProtos(installs),
	}), nil
}
