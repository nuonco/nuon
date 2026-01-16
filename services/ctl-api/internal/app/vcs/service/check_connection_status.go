package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v50/github"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type VCSConnectionStatusResponse struct {
	Status              string                `json:"status"`
	GithubInstallID     string                `json:"github_install_id"`
	Account             *VCSConnectionAccount `json:"account"`
	SuspendedAt         *time.Time            `json:"suspended_at,omitempty"`
	SuspendedBy         *github.User          `json:"suspended_by,omitempty"`
	Permissions         map[string]string     `json:"permissions"`
	RepositorySelection string                `json:"repository_selection"`
	CheckedAt           time.Time             `json:"checked_at"`
	Error               string                `json:"error,omitempty"`
}

type VCSConnectionAccount struct {
	Login string `json:"login"`
	ID    int64  `json:"id"`
	Type  string `json:"type"`
}

// @ID                      CheckVCSConnectionStatus
// @Summary                 check the real-time status of a VCS connection
// @Description.markdown    check_vcs_connection_status.md
// @Param                   connection_id   path    string  true    "connection ID"
// @Tags                    vcs
// @Accept                  json
// @Produce                 json
// @Security                APIKey
// @Security                OrgID
// @Failure                 400 {object}    stderr.ErrResponse
// @Failure                 401 {object}    stderr.ErrResponse
// @Failure                 403 {object}    stderr.ErrResponse
// @Failure                 404 {object}    stderr.ErrResponse
// @Failure                 500 {object}    stderr.ErrResponse
// @Success                 200 {object}    VCSConnectionStatusResponse
// @Router                  /v1/vcs/connections/{connection_id}/check-status [get]
func (s *service) CheckConnectionStatus(ctx *gin.Context) {
	vcsID := ctx.Param("connection_id")

	currentOrg, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	vcsConn, err := s.getConnection(ctx, currentOrg.ID, vcsID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org vcs connection: %w", err))
		return
	}

	statusResp, err := s.checkGithubInstallationStatus(ctx, vcsConn)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to check installation status: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, statusResp)
}

func (s *service) checkGithubInstallationStatus(
	ctx context.Context,
	vcsConn *app.VCSConnection,
) (*VCSConnectionStatusResponse, error) {
	checkedAt := time.Now().UTC()

	ghClient, err := s.helpers.GetJWTVCSConnectionClient()
	if err != nil {
		return nil, fmt.Errorf("unable to create jwt vcs connection client: %w", err)
	}

	installID, err := strconv.ParseInt(vcsConn.GithubInstallID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("unable to convert github install ID to int: %w", err)
	}

	installation, _, err := ghClient.Apps.GetInstallation(ctx, installID)
	if err != nil {
		return &VCSConnectionStatusResponse{
			Status:          "unknown",
			GithubInstallID: vcsConn.GithubInstallID,
			CheckedAt:       checkedAt,
			Error:           fmt.Sprintf("failed to fetch installation from GitHub: %v", err),
		}, nil
	}

	return buildStatusResponse(installation, vcsConn.GithubInstallID, checkedAt), nil
}

func buildStatusResponse(
	installation *github.Installation,
	installIDStr string,
	checkedAt time.Time,
) *VCSConnectionStatusResponse {
	resp := &VCSConnectionStatusResponse{
		GithubInstallID:     installIDStr,
		CheckedAt:           checkedAt,
		RepositorySelection: installation.GetRepositorySelection(),
	}

	if installation.SuspendedAt != nil {
		resp.Status = "suspended"
		resp.SuspendedAt = &installation.SuspendedAt.Time
		if installation.SuspendedBy != nil {
			resp.SuspendedBy = installation.SuspendedBy
		}
	} else {
		resp.Status = "active"
	}

	if installation.Account != nil {
		resp.Account = &VCSConnectionAccount{
			Login: installation.Account.GetLogin(),
			ID:    installation.Account.GetID(),
			Type:  installation.Account.GetType(),
		}
	}

	resp.Permissions = convertInstallationPermissions(installation.Permissions)

	return resp
}

func convertInstallationPermissions(perms *github.InstallationPermissions) map[string]string {
	if perms == nil {
		return make(map[string]string)
	}

	result := make(map[string]string)

	if perms.Actions != nil {
		result["actions"] = *perms.Actions
	}
	if perms.Contents != nil {
		result["contents"] = *perms.Contents
	}
	if perms.Metadata != nil {
		result["metadata"] = *perms.Metadata
	}
	if perms.PullRequests != nil {
		result["pull_requests"] = *perms.PullRequests
	}
	if perms.Workflows != nil {
		result["workflows"] = *perms.Workflows
	}
	if perms.Issues != nil {
		result["issues"] = *perms.Issues
	}
	if perms.Checks != nil {
		result["checks"] = *perms.Checks
	}
	if perms.Deployments != nil {
		result["deployments"] = *perms.Deployments
	}
	if perms.Statuses != nil {
		result["statuses"] = *perms.Statuses
	}

	return result
}
