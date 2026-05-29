package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type AdminSetAdminsRequest struct {
	// EmailOrSubjectOrIDs is a list of accounts to update, identified by email, subject, or account ID.
	EmailOrSubjectOrIDs []string `json:"email_or_subject_or_ids" validate:"required,min=1,dive,required"`
	// IsAdmin is the value to set the is_admin flag to for every listed account.
	IsAdmin bool `json:"is_admin"`
}

func (c *AdminSetAdminsRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						AdminSetAdmins
// @Summary				grant or revoke admin access for accounts.
// @Description.markdown	admin_set_admins.md
// @Param					req	body	AdminSetAdminsRequest	true	"Input"
// @Tags					general/admin
// @Security				AdminEmail
// @Accept					json
// @Produce					json
// @Success				200	{object}	app.EmptyResponse
// @Router					/v1/general/admin-set-admins [POST]
func (s *service) AdminSetAdmins(ctx *gin.Context) {
	var req AdminSetAdminsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	if err := s.setAdmins(ctx, req.EmailOrSubjectOrIDs, req.IsAdmin); err != nil {
		ctx.Error(fmt.Errorf("unable to set admins: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, app.EmptyResponse{})
}

func (s *service) setAdmins(ctx context.Context, emailOrSubjectOrIDs []string, isAdmin bool) error {
	ids := make([]string, 0, len(emailOrSubjectOrIDs))
	for _, identifier := range emailOrSubjectOrIDs {
		acct, err := s.acctClient.FindAccount(ctx, identifier)
		if err != nil {
			return fmt.Errorf("invalid account %q: %w", identifier, err)
		}
		ids = append(ids, acct.ID)
	}

	// Update is_admin explicitly via column name so that setting it to false is not dropped as a zero value.
	res := s.db.WithContext(ctx).
		Model(&app.Account{}).
		Where("id IN ?", ids).
		Update("is_admin", isAdmin)
	if res.Error != nil {
		return fmt.Errorf("unable to update accounts: %w", res.Error)
	}

	return nil
}
