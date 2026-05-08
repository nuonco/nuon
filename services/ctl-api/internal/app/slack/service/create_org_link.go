package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

// CreateOrgLinkRequest binds an existing Slack workspace (TeamID) to the
// calling org. The bound workspace must already have an active
// SlackInstallation; this is enforced at the application layer.
type CreateOrgLinkRequest struct {
	TeamID string `json:"team_id" validate:"required"`
}

func (r *CreateOrgLinkRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(r); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						CreateSlackOrgLink
// @Summary				Bind a Slack workspace to the current org
// @Description			Creates a verified SlackOrgLink between the supplied TeamID and the calling org. Used by the Phase 4 confirmation flow when a user finishes the Slack OAuth round-trip and selects the Nuon org to attach the workspace to.
// @Tags					slack
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					org_id	path	string					true	"Org ID"
// @Param					req		body	CreateOrgLinkRequest	true	"Input"
// @Success				201	{object}	app.SlackOrgLink
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				409	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Router					/v1/orgs/{org_id}/slack/org-links [POST]
func (s *service) CreateOrgLink(ctx *gin.Context) {
	acct, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	req := CreateOrgLinkRequest{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	link, err := s.createOrgLink(ctx, acct, org.ID, req.TeamID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create slack org link: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, link)
}

func (s *service) createOrgLink(ctx context.Context, acct *app.Account, orgID, teamID string) (*app.SlackOrgLink, error) {
	// Ensure the workspace is actually installed and active. We don't have a
	// PG FK for this (soft-delete + FK incompatibility) so it's enforced
	// here at the application layer.
	var install app.SlackInstallation
	res := s.db.WithContext(ctx).
		Where(app.SlackInstallation{
			TeamID: teamID,
			Status: app.SlackInstallationStatusActive,
		}).
		First(&install)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, stderr.ErrNotFound{Err: fmt.Errorf("no active slack installation for team %q", teamID)}
	}
	if res.Error != nil {
		return nil, res.Error
	}

	// Pre-check for an existing live link on (team_id, org_id). The unique
	// index would catch the dup at insert time too, but a generic 500 from
	// a constraint violation is a worse user experience than a clean 409.
	var existing app.SlackOrgLink
	dupRes := s.db.WithContext(ctx).
		Where(app.SlackOrgLink{TeamID: teamID, OrgID: orgID}).
		First(&existing)
	if dupRes.Error == nil {
		return nil, stderr.ErrConflict{
			Err:         fmt.Errorf("slack org link already exists for team %q and org %q", teamID, orgID),
			Description: "this Slack workspace is already linked to this org",
		}
	}
	if !errors.Is(dupRes.Error, gorm.ErrRecordNotFound) {
		return nil, dupRes.Error
	}

	// Set the calling account on the context so SlackOrgLink.BeforeCreate
	// can resolve CreatedByID via the GORM hook.
	ctx = cctx.SetAccountContext(ctx, acct)

	link := &app.SlackOrgLink{
		TeamID:            teamID,
		OrgID:             orgID,
		Status:            app.SlackOrgLinkStatusVerified,
		LinkedByAccountID: acct.ID,
	}
	if err := s.db.WithContext(ctx).Create(link).Error; err != nil {
		return nil, err
	}
	return link, nil
}
