package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateSandboxRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (c *CreateSandboxRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @ID AdminCreateSandbox
// @Summary	create a new sandbox
// @Description.markdown create_sandbox.md
// @Param			req	body	CreateSandboxRequest	true	"Input"
// @Tags			sandboxes/admin
// @Accept			json
// @Produce		json
// @Success		201	{object}	app.Sandbox
// @Router			/v1/sandboxes [post]
func (s *service) CreateSandbox(ctx *gin.Context) {
	req := CreateSandboxRequest{}
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	sandbox, err := s.createSandbox(ctx, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create sandbox: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, sandbox)
}

func (s *service) createSandbox(ctx context.Context, req *CreateSandboxRequest) (*app.Sandbox, error) {
	sandbox := &app.Sandbox{
		Name:        req.Name,
		Description: req.Description,
	}
	if err := s.db.WithContext(ctx).Create(sandbox).Error; err != nil {
		return nil, fmt.Errorf("unable to create sandbox: %w", err)
	}

	return sandbox, nil
}
