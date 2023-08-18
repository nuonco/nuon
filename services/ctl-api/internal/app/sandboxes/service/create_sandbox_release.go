package service

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateSandboxReleaseRequest struct {
	TerraformVersion string `json:"terraform_version,omitempty" validate:"required"`
	Version          string `json:"version,omitempty" validate:"required"`
}

func (c *CreateSandboxReleaseRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @BasePath /v1/sandboxes
// Create a new sandbox
// @Summary create a new sandbox
// @Schemes
// @Description create a new sandbox
// @Param req body CreateSandboxReleaseRequest true "Input"
// @Param sandbox_id path string true "sandbox ID"
// @Tags sandboxes/internal
// @Accept json
// @Produce json
// @Success 201 {object} app.Sandbox
// @Router /v1/sandboxes/{sandbox_id}/release [post]
func (s *service) CreateSandboxRelease(ctx *gin.Context) {
	sandboxID := ctx.Param("sandbox_id")
	req := CreateSandboxReleaseRequest{}
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	sandbox, err := s.createSandboxRelease(ctx, sandboxID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create sandbox: %w", err))
		return
	}
	ctx.JSON(http.StatusCreated, sandbox)
}

func (s *service) createSandboxRelease(ctx context.Context, sandboxID string, req *CreateSandboxReleaseRequest) (*app.SandboxRelease, error) {
	sandbox := app.Sandbox{
		ID: sandboxID,
	}

	// build base URL
	baseURL := s.cfg.SandboxArtifactsBaseURL
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	baseURL += filepath.Join(sandbox.Name, req.Version) + "/"

	// create release
	sandboxRelease := app.SandboxRelease{
		Version:                 req.Version,
		TerraformVersion:        req.TerraformVersion,
		ProvisionPolicyURL:      baseURL + "provision.json",
		TrustPolicyURL:          baseURL + "trust.json",
		DeprovisionPolicyURL:    baseURL + "deprovision.json",
		OneClickRoleTemplateURL: baseURL + "install-role.yaml",
	}

	err := s.db.Model(&sandbox).Association("Releases").Append(&sandboxRelease)
	if err != nil {
		return nil, fmt.Errorf("unable to save release: %w", err)
	}
	return &sandbox.Releases[len(sandbox.Releases)-1], nil
}
