package instances

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	instancev1 "github.com/powertoolsdev/mono/pkg/types/api/instance/v1"
	"github.com/powertoolsdev/mono/services/api/internal/servers/converters"
)

func (s *server) GetInstancesByInstall(
	ctx context.Context,
	req *connect.Request[instancev1.GetInstancesByInstallRequest],
) (*connect.Response[instancev1.GetInstancesByInstallResponse], error) {
	// run protobuf validations
	if err := req.Msg.Validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	instances, err := s.Svc.GetInstancesByInstall(ctx, req.Msg.InstallId)
	if err != nil {
		return nil, fmt.Errorf("unable to get instances: %w", err)
	}

	return connect.NewResponse(&instancev1.GetInstancesByInstallResponse{
		Instances: converters.InstanceModelsToProtos(instances),
	}), nil
}
