package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateSandboxRequest struct {
	Name        string
	Description string
}

// @BasePath /v1/sandboxes
// Create a new sandbox
// @Summary create a new sandbox
// @Schemes
// @Description create a new sandbox
// @Param req body CreateSandboxRequest true "Input"
// @Tags sandboxes/internal
// @Accept json
// @Produce json
// @Success 201 {object} app.Sandbox
// @Router /v1/sandboxes [post]
func (s *service) CreateSandbox(ctx *gin.Context) {
	req := CreateSandboxRequest{}
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
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
