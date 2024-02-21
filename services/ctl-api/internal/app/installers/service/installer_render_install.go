package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type RenderedInstall struct {
	Install          *app.Install      `json:"install"`
	Installer        RenderedInstaller `json:"installer"`
	InstallerContent string            `json:"installer_content"`
}

// @ID RenderInstallerInstall
// @Summary	render an installer install
// @Description.markdown installer_render_install.md
// @Tags installers
// @Accept			json
// @Produce		json
// @Param			installer_slug	path		string	true	"installer slug or ID"
// @Param			install_id		path		string	true	"install id"
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{object}	RenderedInstall
// @Router			/v1/installer/{installer_slug}/install/{install_id}/render [get]
func (s *service) RenderInstallerInstall(ctx *gin.Context) {
	installID := ctx.Param("install_id")
	slugOrID := ctx.Param("installer_slug")
	if slugOrID == "" {
		ctx.Error(fmt.Errorf("slug or id must be set"))
		return
	}

	install, err := s.findInstall(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install %s: %w", installID, err))
		return
	}

	installer, err := s.getAppInstaller(ctx, slugOrID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app installer: %w", err))
		return
	}
	var inputs app.AppInputConfig
	if len(installer.App.AppInputConfigs) > 0 {
		inputs = installer.App.AppInputConfigs[0]
	}

	renderedContent, err := s.renderInstallContent(ctx, installer.Metadata.PostInstallMarkdown, install, installer)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to render install content: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, RenderedInstall{
		Installer: RenderedInstaller{
			App:         installer.App,
			AppInputs:   inputs,
			AppSandbox:  installer.App.AppSandboxConfigs[0],
			SandboxMode: installer.App.Org.SandboxMode,
			Metadata:    installer.Metadata,
		},
		Install:          install,
		InstallerContent: renderedContent,
	})
}

func (s *service) renderInstallContent(ctx context.Context, templateStr string, install *app.Install, installer *app.AppInstaller) (string, error) {
	installJSON, err := json.Marshal(install)
	if err != nil {
		return "", fmt.Errorf("unable to convert install to json: %w", err)
	}

	var installMap map[string]interface{}
	if err := json.Unmarshal(installJSON, &installMap); err != nil {
		return "", fmt.Errorf("unable to convert install to map: %w", err)
	}

	appJSON, err := json.Marshal(install.App)
	if err != nil {
		return "", fmt.Errorf("unable to convert app to json: %w", err)
	}

	var appMap map[string]interface{}
	if err := json.Unmarshal(appJSON, &appMap); err != nil {
		return "", fmt.Errorf("unable to convert app to map: %w", err)
	}

	installerJSON, err := json.Marshal(installer)
	if err != nil {
		return "", fmt.Errorf("unable to convert installer to json: %w", err)
	}

	var installerMap map[string]interface{}
	if err := json.Unmarshal(installerJSON, &installerMap); err != nil {
		return "", fmt.Errorf("unable to convert app to map: %w", err)
	}

	data := map[string]interface{}{
		"nuon": map[string]interface{}{
			"install":   installMap,
			"installer": installerMap,
			"app":       appMap,
		},
	}

	temp, err := template.New("post-install").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("unable to parse template: %w", err)
	}

	buf := new(bytes.Buffer)
	if err := temp.Execute(buf, data); err != nil {
		return "", fmt.Errorf("unable to render template: %w", err)
	}

	return buf.String(), nil
}
