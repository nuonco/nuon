package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/oidc/githubauth"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/require"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type CreateConnectionCallbackRequest struct {
	OrgID           string `json:"org_id" validate:"required"`
	GithubInstallID string `json:"github_install_id" validate:"required"`
	// Code is the GitHub user-authorization code issued after the install.
	// It proves the caller can see github_install_id; the id alone can be
	// spoofed.
	Code string `json:"code"`
}

func (c *CreateConnectionCallbackRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						CreateVCSConnectionCallback
// @Summary					create a vcs connection from the github app install callback
// @Description.markdown	create_connection_callback.md
// @Tags					vcs
// @Security				APIKey
// @Accept					json
// @Produce					json
// @Param					req	body		CreateConnectionCallbackRequest	true	"Input"
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					403	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					409	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					201	{object}	app.VCSConnection
// @Router					/v1/vcs/connection-callback [post]
func (s *service) CreateConnectionCallback(ctx *gin.Context) {
	var req CreateConnectionCallbackRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	acct, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	if s.disableConnectionCallback {
		ctx.Error(stderr.ErrAuthorization{
			Err:         fmt.Errorf("vcs connection callback is disabled"),
			Description: "Connecting a GitHub app is temporarily disabled.",
		})
		return
	}

	if err := acct.AllPermissions.CanPerform(req.OrgID, permissions.PermissionCreate); err != nil {
		ctx.Error(require.PermissionDeniedError(acct, req.OrgID, permissions.PermissionCreate, require.ScopeFromPath(ctx.FullPath())))
		return
	}

	org, err := s.getOrg(ctx, req.OrgID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org: %w", err))
		return
	}

	if err := s.verifyInstallationOwnership(ctx, acct, req); err != nil {
		ctx.Error(err)
		return
	}

	// Fetch org name
	ghAccount, err := s.ghClient.GetInstallationAccount(ctx, req.GithubInstallID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org name: %w", err))
		return
	}
	ghAccountID := strconv.FormatInt(ghAccount.GetID(), 10)
	ghAccountName := ghAccount.GetLogin()

	// Create object
	dbCtx := cctx.SetAccountContext(ctx, acct)
	vcsConn, err := s.createOrgConnection(dbCtx, org.ID, req.GithubInstallID, ghAccountID, ghAccountName)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create org connection: %w", err))
		return
	}

	// Create background health check queue for the connection
	if _, err := s.helpers.CreateConnectionQueue(dbCtx, vcsConn); err != nil {
		s.l.Warn("unable to create vcs connection queue",
			zap.String("vcs_connection_id", vcsConn.ID),
			zap.Error(err),
		)
	}

	ctx.JSON(http.StatusCreated, vcsConn)
}

// verifyInstallationOwnership rejects installation ids the authorizing GitHub
// user cannot see. Without it any account that can create an org could
// attach, read, and uninstall another customer's installation.
func (s *service) verifyInstallationOwnership(ctx *gin.Context, acct *app.Account, req CreateConnectionCallbackRequest) error {
	if s.disableOwnershipCheck {
		s.l.Warn("vcs connection ownership check is disabled; attaching installation without verification",
			zap.String("org_id", req.OrgID),
			zap.String("github_install_id", req.GithubInstallID),
			zap.String("account_id", acct.ID),
		)
		return nil
	}

	if req.Code == "" {
		return stderr.NewInvalidRequest(fmt.Errorf("code is required"))
	}

	err := s.ghClient.VerifyInstallationOwnership(ctx, req.Code, req.GithubInstallID)
	if err == nil {
		return nil
	}

	s.l.Warn("vcs connection ownership check failed",
		zap.String("org_id", req.OrgID),
		zap.String("github_install_id", req.GithubInstallID),
		zap.String("account_id", acct.ID),
		zap.Error(err),
	)
	if errors.Is(err, githubauth.ErrInstallationNotOwned) {
		return stderr.ErrAuthorization{
			Err:         fmt.Errorf("github installation %s is not accessible to the authorizing user", req.GithubInstallID),
			Description: "This GitHub installation does not belong to the GitHub account you authorized.",
		}
	}
	return fmt.Errorf("unable to verify github installation ownership: %w", err)
}

func (s *service) getOrg(ctx context.Context, orgID string) (*app.Org, error) {
	org := app.Org{}
	res := s.db.WithContext(ctx).
		First(&org, "id = ?", orgID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get org %s: %w", orgID, res.Error)
	}

	return &org, nil
}

func (s *service) createOrgConnection(
	ctx context.Context, orgID, githubInstallID, githubAccountID, githubAccountName string,
) (*app.VCSConnection, error) {
	vcsConn := app.VCSConnection{
		OrgID:             orgID,
		GithubInstallID:   githubInstallID,
		GithubAccountID:   githubAccountID,
		GithubAccountName: githubAccountName,
	}

	if err := s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "org_id"}, {Name: "github_install_id"}, {Name: "deleted_at"}},
		DoNothing: true,
	}).Create(&vcsConn).Error; err != nil {
		return nil, fmt.Errorf("unable to create vcs_connection: %w", err)
	}

	// NOTE(jm): when this is a duplicate, the returned ID is not actually valid, as it is set by the create hook in
	// GORM, but then the conflict happens after.
	return &vcsConn, nil
}
