package installs

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	installv1 "github.com/powertoolsdev/mono/pkg/types/api/install/v1"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/powertoolsdev/mono/services/api/internal/servers/converters"
)

func (s *server) UpsertInstall(
	ctx context.Context,
	req *connect.Request[installv1.UpsertInstallRequest],
) (*connect.Response[installv1.UpsertInstallResponse], error) {
	// run protobuf validations
	if err := req.Msg.Validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	if req.Msg.GetAwsSettings() == nil {
		return nil, fmt.Errorf("only AWS settings are currently supported")
	}
	params := models.InstallInput{
		ID:          converters.ToOptionalStr(req.Msg.Id),
		Name:        req.Msg.Name,
		AppID:       req.Msg.AppId,
		CreatedByID: &req.Msg.CreatedById,
		AwsSettings: &models.AWSSettingsInput{
			Region:     converters.ProtoToAwsRegion(req.Msg.GetAwsSettings().Region),
			IamRoleArn: req.Msg.GetAwsSettings().Role,
		},
		OverrideID: converters.ToOptionalStr(req.Msg.OverrideId),
	}

	install, err := s.Svc.UpsertInstall(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("unable to upsert install: %w", err)
	}

	return connect.NewResponse(&installv1.UpsertInstallResponse{
		Install: converters.InstallModelToProto(install),
	}), nil
}
