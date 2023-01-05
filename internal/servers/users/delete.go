package users

import (
	"context"

	"github.com/bufbuild/connect-go"
	userv1 "github.com/powertoolsdev/protos/api/generated/types/user/v1"
)

func (s *server) DeleteUser(
	ctx context.Context,
	req *connect.Request[userv1.DeleteUserRequest],
) (*connect.Response[userv1.DeleteUserResponse], error) {
	res := connect.NewResponse(&userv1.DeleteUserResponse{})
	return res, nil
}
