package builds

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	buildv1 "github.com/powertoolsdev/mono/pkg/types/api/build/v1"
	"github.com/powertoolsdev/mono/services/api/internal/models"
)

func (s *server) CancelBuild(
	ctx context.Context,
	req *connect.Request[buildv1.CancelBuildRequest],
) (*connect.Response[buildv1.CancelBuildResponse], error) {
	// run protobuf validations
	if err := req.Msg.Validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	//verify that build exists
	var build models.Build
	if err := s.db.WithContext(ctx).First(&build, "id = ?", req.Msg.Id).Error; err != nil {
		return nil, fmt.Errorf("retrieve build failed: %w", err)
	}

	// use temporal client to cancel workflow execution
	if err := s.cancelWorkflow(ctx, build.ID); err != nil {
		return nil, fmt.Errorf("cancel build failed: %s", err)
	}

	return connect.NewResponse(&buildv1.CancelBuildResponse{}), nil
}

func (s *server) cancelWorkflow(ctx context.Context, buildID string) error {
	// use temporal client to cancel workflow execution
	if err := s.temporalClient.CancelWorkflowInNamespace(ctx, "builds", buildID, ""); err != nil {
		return err
	}
	return nil
}
