package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	validatorPkg "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/validator"
)

type AdminDeleteAccountRequest struct {
	EmailOrSubjectOrID string `json:"email_or_subject_or_id" validate:"required"`
}

func (c *AdminDeleteAccountRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

//	@ID						AdminCreateAdminDeleteAccount
//	@Summary				delete an account.
//	@Description.markdown	admin_delete_account.md
//	@Param					req	body	AdminDeleteAccountRequest	true	"Input"
//	@Tags					orgs/admin
//	@Security				AdminEmail
//	@Accept					json
//	@Produce				json
//	@Success				201	{string}	ok
//	@Router					/v1/general/admin-delete-account [POST]
func (s *service) AdminDeleteAccount(ctx *gin.Context) {
	var req AdminDeleteAccountRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request input: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	err := s.createAdminDeleteAccount(ctx, req.EmailOrSubjectOrID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to delete account: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, "ok")
}

func (s *service) createAdminDeleteAccount(ctx context.Context, subjectOrEmail string) error {
	acct, err := s.acctClient.FindAccount(ctx, subjectOrEmail)
	if err != nil {
		return fmt.Errorf("invalid account: %w", err)
	}

	var deletedAcct app.Account
	res := s.db.WithContext(ctx).
		Delete(&deletedAcct, "id = ?", acct.ID)
	if res.Error != nil {
		return fmt.Errorf("unable to delete account: %w", res.Error)
	}

	return nil
}
