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
		return ui.PrintError(err)
	}

	view := ui.NewListView()

	deploys, hasMore, err := s.api.GetInstallComponentDeploys(ctx, installID, componentID, &models.GetPaginatedQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return view.Error(err)
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
		return ui.PrintError(err)
	}

	install, err := s.api.GetInstall(ctx, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	componentID, err = lookup.ComponentID(ctx, s.api, install.AppID, componentID)
	if err != nil {
		return ui.PrintError(err)
	}

	if buildID == "" {
		buildID, err = s.resolveComponentDeployBuildID(ctx, installID, componentID)
		if err != nil {
			return ui.PrintError(err)
		}
	}

	req := &models.ServiceCreateInstallDeployRequest{
		BuildID:            buildID,
		DeployDependents:   deployDeps,
		DeployDependencies: deployDependencies,
	}

	aid, err := s.api.CreateInstallDeploy(ctx, installID, req)
	if err != nil {
		return ui.PrintError(err)
	}

	printActionResult(asJSON, fmt.Sprintf("successfully triggered deploy for install %s", aid.ID), actionResult{
		InstallID: installID,
		ID:        aid.ID,
		Status:    "deploy_triggered",
	})
	return nil
}

// resolveComponentDeployBuildID picks a build when --build-id is omitted.
// Prefer the prior deploy's build so a redeploy stays on the same artifact;
// when the component has never been deployed, fall back to its latest build.
func (s *Service) resolveComponentDeployBuildID(ctx context.Context, installID, componentID string) (string, error) {
	if buildID, ok := priorDeployBuildID(s.api.GetInstallComponentLatestDeploy(ctx, installID, componentID)); ok {
		return buildID, nil
	}

	build, err := s.api.GetComponentLatestBuild(ctx, componentID)
	if err != nil {
		return "", fmt.Errorf("could not resolve a build for component %s; pass --build-id explicitly: %w", componentID, err)
	}
	if build == nil || build.ID == "" {
		return "", fmt.Errorf("could not resolve a build for component %s; pass --build-id explicitly", componentID)
	}
	return build.ID, nil
}

func priorDeployBuildID(deploy *models.AppInstallDeploy, err error) (string, bool) {
	if err != nil || deploy == nil || deploy.BuildID == "" {
		return "", false
	}
	return deploy.BuildID, true
}

func (s *Service) DeployCancel(ctx context.Context, installID, deployID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	deploy, err := s.api.GetInstallDeploy(ctx, installID, deployID)
	if err != nil {
		return ui.PrintError(err)
	}

	workflowID := deploy.WorkflowID
	if workflowID == "" {
		workflowID = deploy.InstallWorkflowID
	}
	if workflowID == "" {
		return ui.PrintError(fmt.Errorf("deploy %s has no associated workflow to cancel", deployID))
	}

	if _, err := s.api.CancelWorkflow(ctx, workflowID); err != nil {
		return ui.PrintJSONError(err)
	}

	printActionResult(asJSON, fmt.Sprintf("successfully requested cancellation of deploy %s", deployID), actionResult{
		InstallID:  installID,
		ID:         deployID,
		WorkflowID: workflowID,
		Status:     "cancellation_requested",
	})
	return nil
}
