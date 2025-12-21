package helpers

import (
	"context"
	"fmt"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

// PublicGitVCSConfigRequest represents a request to create a public git VCS configuration
type PublicGitVCSConfigRequest struct {
	Repo       string `validate:"required"`
	Directory  string `validate:"required"`
	Branch     string `validate:"required"`
	PathFilter string
}

// ConnectedGithubVCSConfigRequest represents a request to create a connected GitHub VCS configuration
type ConnectedGithubVCSConfigRequest struct {
	Repo       string `validate:"required"`
	Directory  string `validate:"required"`
	PathFilter string

	Branch string `validate:"required_without=GitRef"`
	GitRef string `validate:"required_without=Branch"`
}

// VCSConfigRequest is an embeddable type for endpoints that accept VCS configuration
type VCSConfigRequest struct {
	PublicGitVCSConfig       *PublicGitVCSConfigRequest       `json:"public_git_vcs_config"`
	ConnectedGithubVCSConfig *ConnectedGithubVCSConfigRequest `json:"connected_github_vcs_config"`
}

// Validate ensures that only one VCS config type is provided
func (b VCSConfigRequest) Validate() error {
	if b.PublicGitVCSConfig != nil && b.ConnectedGithubVCSConfig != nil {
		return stderr.ErrUser{
			Err:         fmt.Errorf("both public and connected github config set"),
			Description: "only one of connected github or public git configs can be set",
		}
	}

	// VCS config is now optional - can be nil for both
	return nil
}

// BuildConnectedGithubVCSConfig creates a ConnectedGithubVCSConfig from the request
func (h *Helpers) BuildConnectedGithubVCSConfig(
	ctx context.Context,
	req *ConnectedGithubVCSConfigRequest,
	org *app.Org,
) (*app.ConnectedGithubVCSConfig, error) {
	if req == nil {
		return nil, nil
	}

	owner, repo, err := h.SplitRepoSlug(req.Repo)
	if err != nil {
		return nil, err
	}

	vcsConnID, err := h.LookupVCSConnection(ctx, owner, repo, org.VCSConnections)
	if err != nil {
		return nil, err
	}

	return &app.ConnectedGithubVCSConfig{
		Repo:            req.Repo,
		RepoName:        repo,
		RepoOwner:       owner,
		Directory:       req.Directory,
		Branch:          req.Branch,
		PathFilter:      req.PathFilter,
		VCSConnectionID: vcsConnID,
	}, nil
}

// BuildPublicGitVCSConfig creates a PublicGitVCSConfig from the request
func (h *Helpers) BuildPublicGitVCSConfig(
	ctx context.Context,
	req *PublicGitVCSConfigRequest,
) (*app.PublicGitVCSConfig, error) {
	if req == nil {
		return nil, nil
	}

	repo := req.Repo
	if strings.HasPrefix("git@", repo) {
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("invalid git clone url"),
			Description: "Please use either a <owner>/<repo> format, or a full https:// public clone url",
		}
	}

	return &app.PublicGitVCSConfig{
		Repo:       repo,
		Directory:  req.Directory,
		Branch:     req.Branch,
		PathFilter: req.PathFilter,
	}, nil
}
