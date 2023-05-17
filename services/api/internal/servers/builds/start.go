package builds

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/powertoolsdev/mono/pkg/common/shortid"
	buildv1 "github.com/powertoolsdev/mono/pkg/types/api/build/v1"
	"github.com/powertoolsdev/mono/pkg/workflows"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	tclient "go.temporal.io/sdk/client"
)

func (s *server) StartBuild(
	ctx context.Context,
	req *connect.Request[buildv1.StartBuildRequest],
) (*connect.Response[buildv1.StartBuildResponse], error) {
	// run protobuf validations
	if err := req.Msg.Validate(); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	// use gorm model to record that we're starting a build
	buildID, err := shortid.NewNanoID("bld")
	if err != nil {
		return nil, err
	}
	build := models.Build{
		Model: models.Model{
			ID: buildID,
		},
		GitRef:      req.Msg.GitRef,
		ComponentID: req.Msg.ComponentId,
		CreatedByID: req.Msg.CreatedById,
	}
	if err := s.db.WithContext(ctx).Create(&build).Error; err != nil {
		return nil, err
	}

	// start build workflow
	opts := tclient.StartWorkflowOptions{
		ID:        buildID,
		TaskQueue: workflows.DefaultTaskQueue,
		// Memo is non-indexed metadata available when listing workflows
		Memo: map[string]interface{}{
			"git-ref":       req.Msg.GitRef,
			"component-id":  req.Msg.ComponentId,
			"created-by-id": req.Msg.CreatedById,
			"started-by":    "api",
		},
	}
	workflowName := "Build"
	args := buildv1.StartBuildRequest{
		GitRef:      req.Msg.GitRef,
		ComponentId: req.Msg.ComponentId,
		CreatedById: req.Msg.CreatedById,
	}
	tclientBuilds, err := tclient.NewClientFromExisting(*s.temporalClient, tclient.Options{Namespace: "builds"})
	if err != nil {
		return nil, fmt.Errorf("error during StartBuild. Cannot create temporal client: %w", err)
	}
	_, err = tclientBuilds.ExecuteWorkflow(ctx, opts, workflowName, &args)
	if err != nil {
		return nil, fmt.Errorf("failed to start build: %w", err)
	}

	return connect.NewResponse(&buildv1.StartBuildResponse{
		Build: build.ToProto(),
	}), nil
}
