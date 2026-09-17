package service

import (
	_ "embed"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const cloudSchemaBaseURL = "https://api.nuon.co"

type AppConfigTemplateType string

const (
	// not used
	AppConfigTemplateTypeAwsECS       AppConfigTemplateType = "aws-ecs"
	AppConfigTemplateTypeAwsECSBYOVPC AppConfigTemplateType = "aws-ecs-byovpc"
	AppConfigTemplateTypeAwsEKS       AppConfigTemplateType = "aws-eks"
	AppConfigTemplateTypeAwsEKSBYOVPC AppConfigTemplateType = "aws-eks-byovpc"
	AppConfigTemplateTypeAzureAKS     AppConfigTemplateType = "azure-aks"

	// flat app template
	AppConfigTemplateTypeFlat AppConfigTemplateType = "flat"

	// with sources app template
	AppConfigTemplateTypeTopLevel          AppConfigTemplateType = "top-level"
	AppConfigTemplateTypeInstaller         AppConfigTemplateType = "installer"
	AppConfigTemplateTypeRunner            AppConfigTemplateType = "runner"
	AppConfigTemplateTypeSandbox           AppConfigTemplateType = "sandbox"
	AppConfigTemplateTypeInputs            AppConfigTemplateType = "inputs"
	AppConfigTemplateTypeTerraform         AppConfigTemplateType = "terraform"
	AppConfigTemplateTypeTerraformInfra    AppConfigTemplateType = "terraformInfra"
	AppConfigTemplateTypeHelm              AppConfigTemplateType = "helm"
	AppConfigTemplateTypeDockerBuild       AppConfigTemplateType = "docker-build"
	AppConfigTemplateTypeJob               AppConfigTemplateType = "job"
	AppConfigTemplateTypeContainerImage    AppConfigTemplateType = "container-image"
	AppConfigTemplateTypeECRContainerImage AppConfigTemplateType = "ecr-container-image"
)

type AppConfigTemplate struct {
	Format   app.AppConfigVersion
	Type     AppConfigTemplateType
	Filename string
	Content  string
}

// @ID						GetAppConfigTemplate
// @Summary				get an app config template
// @Description.markdown	get_app_config_template.md
// @Tags					apps
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					app_id	path		string					true	"app ID"
// @Param					type	query		AppConfigTemplateType	true	"app template type"
// @Failure				400		{object}	stderr.ErrResponse
// @Failure				401		{object}	stderr.ErrResponse
// @Failure				403		{object}	stderr.ErrResponse
// @Failure				404		{object}	stderr.ErrResponse
// @Failure				500		{object}	stderr.ErrResponse
// @Success				201		{object}	AppConfigTemplate
// @Router					/v1/apps/{app_id}/template-config [get]
func (s *service) GetAppConfigTemplate(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")
	app, err := s.findApp(ctx, org.ID, appID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app %s: %w", appID, err))
		return
	}

	configType := ctx.DefaultQuery("type", string(AppConfigTemplateTypeFlat))

	tmpl, err := s.createAppTemplate(app, AppConfigTemplateType(configType))
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, tmpl)
}

func (s *service) schemaBaseURL() string {
	if s.cfg != nil && s.cfg.PublicAPIURL != "" {
		return strings.TrimRight(s.cfg.PublicAPIURL, "/")
	}
	return cloudSchemaBaseURL
}

func (s *service) withSchemaBaseURL(content string) string {
	return strings.ReplaceAll(content, cloudSchemaBaseURL, s.schemaBaseURL())
}

func (s *service) createAppTemplate(currentApp *app.App, typ AppConfigTemplateType) (*AppConfigTemplate, error) {
	nam := fmt.Sprintf("nuon.%s.toml", currentApp.Name)
	switch typ {
	case AppConfigTemplateTypeTopLevel:
		return &AppConfigTemplate{
			Filename: fmt.Sprintf("nuon-template.%s.toml", currentApp.Name),
			Format:   app.AppConfigVersionDefault,
			Content:  s.withSchemaBaseURL(fmt.Sprintf(topLevelConfig, nam, nam)),
		}, nil
	case AppConfigTemplateTypeInstaller:
		return &AppConfigTemplate{
			Filename: "template_installer.toml",
			Format:   app.AppConfigVersionDefault,
			Content:  s.withSchemaBaseURL(installerConfig),
		}, nil
	case AppConfigTemplateTypeRunner:
		return &AppConfigTemplate{
			Filename: "template_runner.toml",
			Format:   app.AppConfigVersionDefault,
			Content:  s.withSchemaBaseURL(runnerConfig),
		}, nil
	case AppConfigTemplateTypeSandbox:
		return &AppConfigTemplate{
			Filename: "template_sandbox.toml",
			Format:   app.AppConfigVersionDefault,
			Content:  s.withSchemaBaseURL(sandboxConfig),
		}, nil
	case AppConfigTemplateTypeInputs:
		return &AppConfigTemplate{
			Filename: "template_inputs.toml",
			Format:   app.AppConfigVersionDefault,
			Content:  s.withSchemaBaseURL(inputsConfig),
		}, nil
	case AppConfigTemplateTypeTerraform:
		return &AppConfigTemplate{
			Filename: "template_terraform_component.toml",
			Format:   app.AppConfigVersionDefault,
			Content:  s.withSchemaBaseURL(terraformComponentConfig),
		}, nil
	case AppConfigTemplateTypeTerraformInfra:
		return &AppConfigTemplate{
			Filename: "template_terraform_infra_component.toml",
			Format:   app.AppConfigVersionDefault,
			Content:  s.withSchemaBaseURL(terraformInfraComponentConfig),
		}, nil
	case AppConfigTemplateTypeHelm:
		return &AppConfigTemplate{
			Filename: "template_helm_component.toml",
			Format:   app.AppConfigVersionDefault,
			Content:  s.withSchemaBaseURL(helmComponentConfig),
		}, nil
	case AppConfigTemplateTypeDockerBuild:
		return &AppConfigTemplate{
			Filename: "template_docker_build_component.toml",
			Format:   app.AppConfigVersionDefault,
			Content:  s.withSchemaBaseURL(dockerBuildComponentConfig),
		}, nil
	case AppConfigTemplateTypeContainerImage:
		return &AppConfigTemplate{
			Filename: "template_container_image_component.toml",
			Format:   app.AppConfigVersionDefault,
			Content:  s.withSchemaBaseURL(containerImageComponentConfig),
		}, nil
	case AppConfigTemplateTypeJob:
		return &AppConfigTemplate{
			Filename: "template_job_component.toml",
			Format:   app.AppConfigVersionDefault,
			Content:  s.withSchemaBaseURL(jobComponentConfig),
		}, nil
	case AppConfigTemplateTypeECRContainerImage:
		return &AppConfigTemplate{
			Filename: "template_ecr_container_image_component.toml",
			Format:   app.AppConfigVersionDefault,
			Content:  s.withSchemaBaseURL(ecrContainerImageComponentConfig),
		}, nil
	default:
		return &AppConfigTemplate{
			Filename: fmt.Sprintf("nuon-template.%s.toml", currentApp.Name),
			Format:   app.AppConfigVersionDefault,
			Content:  s.withSchemaBaseURL(fmt.Sprintf(flatAppConfigTemplate, nam, nam, currentApp.Name)),
		}, nil
	}
}
