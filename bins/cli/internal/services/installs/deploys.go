package installs

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/pkg/oci/imageref"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) ComponentDeploysList(ctx context.Context, installID, componentID string, offset, limit int, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return err
	}

	view := ui.NewListView()

	deploys, hasMore, err := s.api.GetInstallComponentDeploys(ctx, installID, componentID, &models.GetPaginatedQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return err
	}

	if asJSON {
		ui.PrintJSON(deploys)
		return nil
	}

	data := [][]string{
		{
			"ID",
			"STATUS",
			"TYPE",
			"BUILD ID",
			"IMAGE",
			"CREATED AT",
			"COMPONENT CONFIG VERSION",
		},
	}
	for _, deploy := range deploys {
		var image string
		if b := deploy.ComponentBuild; b != nil && b.SourceDigest != "" {
			image = imageref.DisplayRef(imageref.Source{
				SourceImage:  b.SourceImage,
				SourceRef:    b.SourceRef,
				ResolvedTag:  b.ResolvedTag,
				SourceDigest: b.SourceDigest,
			})
		}
		data = append(data, []string{
			deploy.ID,
			deploy.Status,
			string(deploy.InstallDeployType),
			deploy.BuildID,
			image,
			deploy.CreatedAt,
			fmt.Sprintf("%d", deploy.ComponentConfigVersion),
		})
	}
	view.RenderPaging(data, offset, limit, hasMore)
	return nil
}

func (s *Service) ComponentDeployCreate(ctx context.Context, installID, componentID, buildID string, deployDeps, deployDependencies, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return err
	}

	if buildID == "" {
		latest, err := s.api.GetInstallComponentLatestDeploy(ctx, installID, componentID)
		if err != nil {
			return err
		}
		if latest == nil || latest.BuildID == "" {
			return ui.PrintError(fmt.Errorf("could not resolve a build for component %s; pass --build-id explicitly", componentID))
		}
		buildID = latest.BuildID
	}

	req := &models.ServiceCreateInstallDeployRequest{
		BuildID:            buildID,
		DeployDependents:   deployDeps,
		DeployDependencies: deployDependencies,
	}

	aid, err := s.api.CreateInstallDeploy(ctx, installID, req)
	if err != nil {
		return err
	}

	printActionResult(asJSON, fmt.Sprintf("successfully triggered deploy for install %s", aid.ID), actionResult{
		InstallID: installID,
		ID:        aid.ID,
		Status:    "deploy_triggered",
	})
	return nil
}

func (s *Service) DeployCancel(ctx context.Context, installID, deployID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return err
	}

	deploy, err := s.api.GetInstallDeploy(ctx, installID, deployID)
	if err != nil {
		return err
	}

	workflowID := deploy.WorkflowID
	if workflowID == "" {
		workflowID = deploy.InstallWorkflowID
	}
	if workflowID == "" {
		return ui.PrintError(fmt.Errorf("deploy %s has no associated workflow to cancel", deployID))
	}

	if _, err := s.api.CancelWorkflow(ctx, workflowID); err != nil {
		return err
	}

	printActionResult(asJSON, fmt.Sprintf("successfully requested cancellation of deploy %s", deployID), actionResult{
		InstallID:  installID,
		ID:         deployID,
		WorkflowID: workflowID,
		Status:     "cancellation_requested",
	})
	return nil
}
