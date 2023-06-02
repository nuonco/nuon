package builds

import (
	"context"
	"fmt"

	"github.com/bufbuild/connect-go"
	"github.com/powertoolsdev/mono/pkg/common/shortid"
	apibuildv1 "github.com/powertoolsdev/mono/pkg/types/api/build/v1"
	workflowbuildv1 "github.com/powertoolsdev/mono/pkg/types/workflows/builds/v1"
	wfc "github.com/powertoolsdev/mono/pkg/workflows/client"
	"github.com/powertoolsdev/mono/services/api/internal/models"
	tclient "go.temporal.io/sdk/client"
)

func (s *server) StartBuild(
	ctx context.Context,
	req *connect.Request[apibuildv1.StartBuildRequest],
) (*connect.Response[apibuildv1.StartBuildResponse], error) {
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
	if err = s.db.WithContext(ctx).Create(&build).Error; err != nil {
		return nil, err
	}

	// fetch org and app IDs for workflow
	component := models.Component{}
	s.db.Model(&build).Association("Component").Find(&component)
	app := models.App{}
	s.db.Model(&component).Association("App").Find(&app)
	org := models.Org{}
	s.db.Model(&app).Association("Org").Find(&org)

	// start build workflow
	workflow := "Build"
	namespace := "builds"
	opts := tclient.StartWorkflowOptions{
		ID:        buildID,
		TaskQueue: wfc.APITaskQueue,
		// Memo is non-indexed metadata available when listing workflows
		Memo: map[string]interface{}{
			"git-ref":       req.Msg.GitRef,
			"component-id":  req.Msg.ComponentId,
			"created-by-id": req.Msg.CreatedById,
			"started-by":    "api",
			"org-id":        org.ID,
			"app-id":        app.ID,
		},
	}
	args := workflowbuildv1.BuildRequest{
		BuildId:     buildID,
		GitRef:      req.Msg.GitRef,
		ComponentId: req.Msg.ComponentId,
		OrgId:       org.ID,
		AppId:       app.ID,
	}
	_, err = s.temporalClient.ExecuteWorkflowInNamespace(ctx, namespace, opts, workflow, &args)
	if err != nil {
		return nil, fmt.Errorf("failed to start build: %w", err)
	}

	return connect.NewResponse(&apibuildv1.StartBuildResponse{
		Build: build.ToProto(),
	}), nil
}
