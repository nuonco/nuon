package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CheckVCSConnectionHealthRequest struct {
	VCSConnectionID string `validate:"required"`
}

type CheckVCSConnectionHealthResponse struct {
	Status      app.Status     `json:"status"`
	Description string         `json:"description"`
	RepoCount   int            `json:"repo_count"`
	Metadata    map[string]any `json:"metadata"`
}

// @temporal-gen-v2 activity
func (a *Activities) CheckVCSConnectionHealth(ctx context.Context, req CheckVCSConnectionHealthRequest) (*CheckVCSConnectionHealthResponse, error) {
	var vcsConn app.VCSConnection
	if err := a.db.WithContext(ctx).First(&vcsConn, "id = ?", req.VCSConnectionID).Error; err != nil {
		return nil, fmt.Errorf("unable to get vcs connection: %w", err)
	}

	metadata := map[string]any{
		"github_install_id":   vcsConn.GithubInstallID,
		"github_account_name": vcsConn.GithubAccountName,
		"checked_at":          time.Now().UTC().Format(time.RFC3339),
	}

	// Check GitHub installation status
	installation, err := a.ghClient.GetInstallation(ctx, vcsConn.GithubInstallID)
	if err != nil {
		return &CheckVCSConnectionHealthResponse{
			Status:      app.StatusError,
			Description: fmt.Sprintf("failed to fetch github installation: %v", err),
			Metadata:    metadata,
		}, nil
	}

	if installation.SuspendedAt != nil {
		metadata["suspended_at"] = installation.SuspendedAt.Time.Format(time.RFC3339)
		return &CheckVCSConnectionHealthResponse{
			Status:      app.StatusError,
			Description: "github app installation is suspended",
			Metadata:    metadata,
		}, nil
	}

	// List repos to verify access
	repos, err := a.ghClient.ListInstallationRepos(ctx, &vcsConn)
	if err != nil {
		return &CheckVCSConnectionHealthResponse{
			Status:      app.StatusError,
			Description: fmt.Sprintf("failed to list repos: %v", err),
			Metadata:    metadata,
		}, nil
	}

	metadata["repo_count"] = len(repos)
	metadata["repository_selection"] = installation.GetRepositorySelection()

	return &CheckVCSConnectionHealthResponse{
		Status:      app.StatusSuccess,
		Description: fmt.Sprintf("healthy: %d repos accessible", len(repos)),
		RepoCount:   len(repos),
		Metadata:    metadata,
	}, nil
}
