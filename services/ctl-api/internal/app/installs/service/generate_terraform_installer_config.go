package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @ID						GenerateTerraformInstallerConfig
// @Summary				generate a Terraform installer config
// @Description.markdown	generate_terraform_installer_config.md
// @Param					install_id		path	string	true	"install ID"
// @Tags					installs
// @Accept					json
// @Produce				application/octet-stream
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{file}	string
// @Router					/v1/installs/{install_id}/generate-terraform-installer-config [get]
func (s *service) GenerateTerraformInstallerConfig(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	installCfg, err := s.genTerraformInstallerConfig(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("error generating config from current state: %w", err))
		return
	}

	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.tf\"", "terraform.tfvars"))
	ctx.Data(http.StatusOK, "application/octet-stream", []byte(installCfg))
}

func (s *service) genTerraformInstallerConfig(ctx context.Context, installID string) (string, error) {
	install, err := s.getInstall(ctx, installID)
	if err != nil {
		return "", err
	}
	runnerGroup, err := s.getInstallRunnerGroup(ctx, installID)
	if err != nil {
		return "", err
	}
	if len(runnerGroup.Runners) == 0 {
		return "", fmt.Errorf("no runners in install runner group")
	}

	token, err := s.runnersHelpers.CreateToken(ctx, runnerGroup.Runners[0].ID)
	if err != nil {
		return "", err
	}

	phoneHomeID := install.AWSAccount.ID
	phoneHomeURL := fmt.Sprintf(
		"%s/v1/installs/%s/phone-home/%s",
		s.cfg.PublicAPIURL,
		install.ID,
		phoneHomeID,
	)

	file := fmt.Sprintf(`# Terraform variables file for creating an install.
# All required values included.

nuon_org_id = "%s"
nuon_app_id = "%s"
nuon_install_id = "%s"
nuon_api_token = "%s"
phone_home_url = "%s"`,
		install.OrgID, install.AppID, install.ID, token.Token, phoneHomeURL)

	return file, nil
}
