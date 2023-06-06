package orgs

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	orgv1 "github.com/powertoolsdev/mono/pkg/types/api/org/v1"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	"github.com/powertoolsdev/mono/services/api/internal/servers/converters"
)

func (s *server) UpsertOrg(
	ctx context.Context,
	req *connect.Request[orgv1.UpsertOrgRequest],
) (*connect.Response[orgv1.UpsertOrgResponse], error) {
	// run protobuf validations
	if err := req.Msg.Validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	org, err := s.Svc.UpsertOrg(ctx, models.OrgInput{
		ID:              converters.ToOptionalStr(req.Msg.Id),
		Name:            req.Msg.Name,
		OwnerID:         req.Msg.OwnerId,
		GithubInstallID: converters.ToOptionalStr(req.Msg.GithubInstallId),
		OverrideID:      converters.ToOptionalStr(req.Msg.OverrideId),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to upsert org: %w", err)
	}

	return connect.NewResponse(&orgv1.UpsertOrgResponse{
		Org: converters.OrgModelToProto(org),
	}), nil
}
