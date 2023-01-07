package installs

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	installv1 "github.com/powertoolsdev/protos/api/generated/types/install/v1"
)

func (s *server) DeleteInstall(
	ctx context.Context,
	req *connect.Request[installv1.DeleteInstallRequest],
) (*connect.Response[installv1.DeleteInstallResponse], error) {
	deleted, err := s.Svc.DeleteInstall(ctx, req.Msg.Id)
	if err != nil {
		return nil, fmt.Errorf("unable to delete install: %w", err)
	}

	return connect.NewResponse(&installv1.DeleteInstallResponse{
		Deleted: deleted,
	}), nil
}
