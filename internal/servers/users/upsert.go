package users

import (
	"context"

	"github.com/bufbuild/connect-go"
	userv1 "github.com/powertoolsdev/protos/api/generated/types/user/v1"
)

func (s *server) UpsertUser(
	ctx context.Context,
	req *connect.Request[userv1.UpsertUserRequest],
) (*connect.Response[userv1.UpsertUserResponse], error) {
	res := connect.NewResponse(&userv1.UpsertUserResponse{})
	return res, nil
}

func (s *server) UpsertUserOrg(
	ctx context.Context,
	req *connect.Request[userv1.UpsertUserOrgRequest],
) (*connect.Response[userv1.UpsertUserOrgResponse], error) {
	res := connect.NewResponse(&userv1.UpsertUserOrgResponse{})
	return res, nil
}
