package installs

import (
	"context"
	"fmt"
	"time"

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

func (s *Service) ComponentDeployCreate(ctx context.Context, installID, componentID, buildID string, deployDeps, deployDependencies bool, requestID string, asJSON bool) error {
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
		buildID, err = s.resolveComponentDeployBuildID(ctx, install, componentID)
		if err != nil {
			return ui.PrintError(err)
		}
	}

	req := &models.ServiceCreateInstallDeployRequest{
		BuildID:            buildID,
		DeployDependents:   deployDeps,
		DeployDependencies: deployDependencies,
		RequestID:          requestID,
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

const (
	deployBuildPageSize = 50
	deployBuildMaxPages = 5
)

// resolveComponentDeployBuildID picks the build the dashboard preselects: the newest active,
// non-preview build on the install's app branch.
func (s *Service) resolveComponentDeployBuildID(ctx context.Context, install *models.AppInstall, componentID string) (string, error) {
	for page := 0; page < deployBuildMaxPages; page++ {
		builds, hasMore, err := s.api.GetComponentBuilds(ctx, componentID, install.AppID, &models.GetPaginatedQuery{
			Offset: page * deployBuildPageSize,
			Limit:  deployBuildPageSize,
		})
		if err != nil {
			return "", fmt.Errorf("could not list builds for component %s; pass --build-id explicitly: %w", componentID, err)
		}
		if buildID, ok := latestDeployableBuildID(builds, install.AppBranchID); ok {
			return buildID, nil
		}
		if !hasMore {
			break
		}
	}
	return "", fmt.Errorf("no active build found for component %s on the install's app branch; pass --build-id explicitly", componentID)
}

func latestDeployableBuildID(builds []*models.AppComponentBuild, appBranchID string) (string, bool) {
	var latest *models.AppComponentBuild
	var latestAt time.Time
	for _, b := range builds {
		if b == nil || b.ID == "" || b.IsPreview {
			continue
		}
		if b.StatusV2 == nil || b.StatusV2.Status != models.AppStatusActive {
			continue
		}
		if appBranchID != "" && buildAppBranchID(b) != appBranchID {
			continue
		}
		createdAt, err := time.Parse(time.RFC3339Nano, b.CreatedAt)
		if err != nil {
			continue
		}
		if latest == nil || createdAt.After(latestAt) {
			latest, latestAt = b, createdAt
		}
	}
	if latest == nil {
		return "", false
	}
	return latest.ID, true
}

func buildAppBranchID(b *models.AppComponentBuild) string {
	if b.AppBranchID != "" {
		return b.AppBranchID
	}
	if b.AppBranchRun != nil && b.AppBranchRun.AppBranch != nil {
		return b.AppBranchRun.AppBranch.ID
	}
	return ""
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
