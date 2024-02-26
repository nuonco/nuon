package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	orgmiddleware "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
)

type AppConfigTemplateType string

const (
	AppConfigTemplateTypeAwsECS       AppConfigTemplateType = "aws-ecs"
	AppConfigTemplateTypeAwsECSBYOVPC AppConfigTemplateType = "aws-ecs-byovpc"
	AppConfigTemplateTypeAwsEKS       AppConfigTemplateType = "aws-eks"
	AppConfigTemplateTypeAwsEKSBYOVPC AppConfigTemplateType = "aws-eks-byovpc"
)

type AppConfigTemplate struct {
	Format   app.AppConfigFmt
	Type     AppConfigTemplateType
	Filename string
	Content  string
}

// @ID GetAppConfigTemplate
// @Summary	get an app config template
// @Description.markdown	get_app_config_template.md
// @Tags			apps
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Param	app_id	path	string				true	"app ID"
// @Param  type query AppConfigTemplateType true "app template type"
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		201				{object}	AppConfigTemplate
// @Router			/v1/apps/{app_id}/template-config [get]
func (s *service) GetAppConfigTemplate(ctx *gin.Context) {
	org, err := orgmiddleware.FromContext(ctx)
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

	tmpl, err := s.createAppTemplate(ctx, app, AppConfigTemplateTypeAwsECS)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, tmpl)
}

func (s *service) createAppTemplate(ctx context.Context, currentApp *app.App, typ AppConfigTemplateType) (*AppConfigTemplate, error) {
	return &AppConfigTemplate{
		Filename: fmt.Sprintf("nuon.%s.toml", currentApp.Name),
		Format:   app.AppConfigFmtToml,
		Content: `version = "v1"

[installer]
name = "installer"
description = "one click installer"
slug = "installer-abc-test"
documentation_url = "docs-url"
community_url = "community-url"
homepage_url = "homepage-url"
github_url = "github-url"
logo_url = "logo-url"
demo_url = "demo url"

[runner]
runner_type = "aws-ecs"

[[runner.env_var]]
name = "runner-env-var"
value = "runner-env-var"

[sandbox]
terraform_version = "1.5.4"
[sandbox.public_repo]
directory = "aws-ecs-byo-vpc"
repo = "nuonco/sandboxes"
branch = "main"

[[sandbox.var]]
name = "vpc_id"
value = "{{.nuon.install.inputs.vpc_id}}"

[inputs]
[[inputs.input]]
name = "vpc_id"
description = "vpc_id to install application into"
default = ""
sensitive = false
display_name = "VPC ID"

[[inputs.input]]
name = "api_key"
description = "API key"
default = ""
sensitive = true
display_name = "API Key"

[[components]]
name = "toml_terraform"
type = "terraform_module"
terraform_version = "1.5.3"

[components.connected_repo]
directory = "infra"
repo = "powertoolsdev/mono"
branch = "main"

[[components.var]]
name = "AWS_REGION"
value = "{{.nuon.install.sandbox.account.region}}"

[[components.var]]
name = "ACCOUNT_ID"
value = "{{.nuon.install.sandbox.account.id}}"

[[components]]
name = "toml_infra"
type = "terraform_module"
terraform_version = "1.5.4"

[components.connected_repo]
directory = "deployment"
repo = "powertoolsdev/mono"
branch = "main"

[[components.var]]
name = "iam_role"
value = "{{.nuon.components.infra.outputs.iam_role}}"

[[components]]
name = "toml_helm"
type = "helm_chart"
chart_name = "e2e-helm"

[components.connected_repo]
directory = "deployment"
repo = "powertoolsdev/mono"
branch = "main"

[[components.value]]
name = "api.ingresses.public_domain"
value = "{{.nuon.components.infra.outputs.iam_role}}"

[[components]]
name = "toml_docker_build"
type = "docker_build"

dockerfile = "Dockerfile"

[components.connected_repo]
directory = "deployment"
repo = "powertoolsdev/mono"
branch = "main"

[[components]]
name = "toml_job"
type = "job"

image_url = "{{.nuon.components.e2e_docker_build.image.repository.uri}}"
tag	  = "{{.nuon.components.e2e_docker_build.image.tag}}"
cmd	  = ["printenv"]
args	  = [""]

[[components.env_var]]
name = "PUBLIC_DOMAIN"
value = "{{.nuon.components.infra.outputs.iam_role}}"

[[components]]
name = "toml_container_image"
type = "container_image"

[components.public]
image_url = "kennethreitz/httpbin"
tag = "latest"

[[components]]
name = "toml_container_image_ecr"
type = "container_image"

[components.aws_ecr]
iam_role_arn = "iam_role_arn"
image_url = "ecr-url"
tag = "latest"
region = "us-west-2"
`,
	}, nil
}
