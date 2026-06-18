package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type FindVCSConnectionsByInstallIDRequest struct {
	GithubInstallID string `json:"github_install_id" validate:"required"`
}

type FindVCSConnectionsByInstallIDResponse struct {
	VCSConnections []app.VCSConnection `json:"vcs_connections"`
}

// @temporal-gen-v2 activity
func (a *Activities) FindVCSConnectionsByInstallID(ctx context.Context, req FindVCSConnectionsByInstallIDRequest) (*FindVCSConnectionsByInstallIDResponse, error) {
	var conns []app.VCSConnection
	if err := a.db.WithContext(ctx).
		Where(app.VCSConnection{GithubInstallID: req.GithubInstallID}).
		Find(&conns).Error; err != nil {
		return nil, fmt.Errorf("unable to find vcs connections by github install id: %w", err)
	}

	return &FindVCSConnectionsByInstallIDResponse{
		VCSConnections: conns,
	}, nil
}
