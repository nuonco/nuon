package installs

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/powertoolsdev/api/internal/models"
	"github.com/powertoolsdev/api/internal/servers/converters"
	installv1 "github.com/powertoolsdev/protos/api/generated/types/install/v1"
)

func (s *server) UpsertInstall(
	ctx context.Context,
	req *connect.Request[installv1.UpsertInstallRequest],
) (*connect.Response[installv1.UpsertInstallResponse], error) {
	if req.Msg.GetAwsSettings() == nil {
		return nil, fmt.Errorf("only AWS settings are currently supported")
	}
	params := models.InstallInput{
		ID:    converters.ToOptionalStr(req.Msg.Id),
		Name:  req.Msg.Name,
		AppID: req.Msg.AppId,
		AwsSettings: &models.AWSSettingsInput{
			Region:     converters.ProtoToAwsRegion(req.Msg.GetAwsSettings().Region),
			IamRoleArn: req.Msg.GetAwsSettings().Role,
		},
	}

	install, err := s.Svc.UpsertInstall(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("unable to upsert install: %w", err)
	}

	return connect.NewResponse(&installv1.UpsertInstallResponse{
		Install: converters.InstallModelToProto(install),
	}), nil
}
