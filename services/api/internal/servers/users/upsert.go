package users

import (
	"context"

	"github.com/bufbuild/connect-go"
	userv1 "github.com/powertoolsdev/mono/pkg/types/api/user/v1"
	"github.com/powertoolsdev/mono/services/api/internal/models"
)

func (s *server) UpsertOrgMember(
	ctx context.Context,
	req *connect.Request[userv1.UpsertOrgMemberRequest],
) (*connect.Response[userv1.UpsertOrgMemberResponse], error) {
	// run protobuf validations
	// TODO 174 temporarily disable validations until migration to shortIDs is complete
	// if err := req.Msg.Validate(); err != nil {
	// 	return nil, fmt.Errorf("input validation failed: %w", err)
	// }

	userOrg, err := s.Svc.UpsertUserOrg(ctx, models.UserOrgInput{
		UserID: req.Msg.UserId,
		OrgID:  req.Msg.OrgId,
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&userv1.UpsertOrgMemberResponse{
		OrgMember: &userv1.OrgMember{
			UserId: userOrg.UserID,
			OrgId:  userOrg.OrgID,
		},
	}), nil
}
