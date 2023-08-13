package service

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type CreateSandboxReleaseRequest struct {
	TerraformVersion string
	Version          string
}

// @BasePath /v1/sandboxes
// Create a new sandbox
// @Summary create a new sandbox
// @Schemes
// @Description create a new sandbox
// @Param req body CreateSandboxReleaseRequest true "Input"
// @Param sandbox_id path string sandbox_id "sandbox ID"
// @Tags sandboxes/internal
// @Accept json
// @Produce json
// @Success 201 {object} app.Sandbox
// @Router /v1/sandboxes/{sandbox_id}/release [post]
func (s *service) CreateSandboxRelease(ctx *gin.Context) {
	sandboxID := ctx.Param("sandbox_id")
	if sandboxID == "" {
		ctx.Error(fmt.Errorf("sandbox_id must be passed in"))
		return
	}

	req := CreateSandboxReleaseRequest{}
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
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
	sandbox, err := s.getSandbox(ctx, sandboxID)
	if err != nil {
		return nil, err
	}

	baseURL := filepath.Join(s.cfg.SandboxArtifactsBaseURL, sandbox.Name, req.Version)
	sandbox.Releases = append(sandbox.Releases, app.SandboxRelease{
		Version:                 req.Version,
		TerraformVersion:        req.TerraformVersion,
		ProvisionPolicyURL:      filepath.Join(baseURL, "provision.json"),
		TrustPolicyURL:          filepath.Join(baseURL, "trust.json"),
		DeprovisionPolicyURL:    filepath.Join(baseURL, "deprovision.json"),
		OneClickRoleTemplateURL: filepath.Join(baseURL, "install-role.yaml"),
	})

	if err := s.db.Save(sandbox).Error; err != nil {
		return nil, fmt.Errorf("unable to save release: %w", err)
	}
	return &sandbox.Releases[len(sandbox.Releases)-1], nil
}
