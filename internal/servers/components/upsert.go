package components

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/powertoolsdev/api/internal/models"
	"github.com/powertoolsdev/api/internal/servers/converters"
	componentv1 "github.com/powertoolsdev/protos/api/generated/types/component/v1"
)

func (s *server) UpsertComponent(
	ctx context.Context,
	req *connect.Request[componentv1.UpsertComponentRequest],
) (*connect.Response[componentv1.UpsertComponentResponse], error) {
	// run protobuf validations
	if err := req.Msg.Validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	params := models.ComponentInput{
		AppID:       req.Msg.AppId,
		ID:          converters.ToOptionalStr(req.Msg.Id),
		Name:        req.Msg.Name,
		CreatedByID: req.Msg.CreatedById,

		// NOTE: the following parameters will not be used once we migrate to the new component ref
		BuildImage: req.Msg.BuildImage,
		Type:       converters.ProtoToComponentType(req.Msg.ComponentType),
	}

	if req.Msg.GetGithubConfig() != nil {
		params.GithubConfig = &models.GithubConfigInput{
			Repo:      req.Msg.GetGithubConfig().Repo,
			Branch:    &req.Msg.GetGithubConfig().Branch,
			RepoOwner: &req.Msg.GetGithubConfig().RepoOwner,
			Directory: &req.Msg.GetGithubConfig().Directory,
		}
	}

	component, err := s.Svc.UpsertComponent(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("unable to upsert component: %w", err)
	}

	return connect.NewResponse(&componentv1.UpsertComponentResponse{
		Component: converters.ComponentModelToProto(component),
	}), nil
}
