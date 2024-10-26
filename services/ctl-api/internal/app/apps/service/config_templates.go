package service

const flatAppConfigTemplate = `# This file contains template values for common Nuon application configuration options.
# To use it for your app, edit as needed, then rename this file to %s and run
#
#   nuon apps sync -c %s
#
# See https://docs.nuon.co/concepts/apps for more information.

version = "v1"
description = " nuon sample app"
display_name = "sample-app"
slack_webhook_url = "https://slack.nuon.co"

[installer]
name               = "My ECS App"
description        = "A demo app that runs on ECS."
documentation_url  = "https://docs.nuon.co/"
community_url      = "https://join.slack.com/t/nuoncommunity/shared_invite/zt-1q323vw9z-C8ztRP~HfWjZx6AXi50VRA"
github_url         = "https://github.com/nuonco"
homepage_url       = "https://www.nuon.co/"
demo_url           = "https://www.nuon.co/"
logo_url           = "https://assets-global.website-files.com/62a2c1332b518a9eedc6de2f/651df2030c43865b9b16046b_Group.png"
og_image_url       = "https://assets-global.website-files.com/62a2c1332b518a9eedc6de2f/651df2030c43865b9b16046b_Group.png"
favicon_url        = "https://assets-global.website-files.com/62a2c1332b518a9eedc6de2f/651df2030c43865b9b16046b_Group.png"
copyright_markdown = """
© 2024 Nuon.
"""
footer_markdown = """
[Terms of Service](https://nuon.co/terms)
"""
post_install_markdown = """
# My ECS App

My ECS App is being deployed.
"""
apps = ["%s"]

[sandbox]
terraform_version = "1.5.4"
[sandbox.public_repo]
directory = "aws-ecs"
repo = "nuonco/sandboxes"
branch = "main"

[runner]
runner_type = "aws-ecs"

[[components]]
name   = "ecs_service"
type = "terraform_module"
terraform_version = "1.5.3"
[components.public_repo]
repo      = "nuonco/guides"
directory = "aws-ecs-tutorial/components/ecs-service"
branch    = "main"
[components.vars]
service_name = "{{.nuon.install.inputs.service_name}}"
cluster_arn = "{{.nuon.install.sandbox.outputs.ecs_cluster.arn}}"
image_url = "{{.nuon.components.docker_image.image.repository.uri}}"
image_tag = "{{.nuon.components.docker_image.image.tag}}"
app_id = "{{.nuon.app.id}}"
org_id = "{{.nuon.org.id}}"
install_id = "{{.nuon.install.id}}"
vpc_id = "{{.nuon.install.sandbox.outputs.vpc.id}}"
domain_name = "api.{{.nuon.install.sandbox.outputs.public_domain.name}}"
zone_id = "{{.nuon.install.sandbox.outputs.public_domain.zone_id}}"
`

const topLevelConfig = `#:schema https://api.nuon.co/v1/general/config-schema
# This file contains template values for common Nuon application configuration options.
# To use it for your app, edit as needed, then rename this file to %s and run
#
#   nuon apps sync -c %s
#
# See https://docs.nuon.co/concepts/apps for more information.

version = "v1"
description = " template with sources"
display_name = "template-app"
slack_webhook_url = "https://slack.nuon.co"

[installer]
source = "template_installer.toml"

[runner]
source = "template_runner.toml"

[sandbox]
source = "template_sandbox.toml"

[inputs]
source = "template_inputs.toml"

[[components]]
source = "template_terraform_component.toml"

[[components]]
source = "template_terraform_infra_component.toml"

[[components]]
source = "template_helm_component.toml"

[[components]]
source = "template_docker_build_component.toml"

[[components]]
source = "template_container_image_component.toml"

[[components]]
source = "template_job_component.toml"

[[components]]
source = "template_ecr_container_image_component.toml"
`

const installerConfig = `#:schema https://api.nuon.co/v1/general/config-schema?source=installer
name = "installer"
description = "one click installer"
documentation_url = "docs-url"
community_url = "community-url"
homepage_url = "homepage-url"
github_url = "github-url"
logo_url = "logo-url"
demo_url = "demo url"
favicon_url = "favicon url"

# optional fields
og_image_url = "og_image url"
post_install_markdown = ""
copyright_markdown = ""
footer_markdown = ""`

const runnerConfig = `#:schema https://api.nuon.co/v1/general/config-schema?source=runner
runner_type = "aws-ecs"

[env_vars]
runner-env-var = "runner-env-var"
`

const sandboxConfig = `#:schema https://api.nuon.co/v1/general/config-schema?source=sandbox
terraform_version = "1.5.4"

# https://docs.nuon.co/guides/install-access-delegation#setup-delegation
# if you are using delegation, otherwise remove
# govcloud clients must reach out for additional configuration
aws_delegation_iam_role_arn = "arn:aws:iam::xxxxxxxxxxxx:role/nuon-aws-ecs-install-access"

[public_repo]
directory = "aws-ecs-byovpc"
repo = "nuonco/sandboxes"
branch = "main"

[vars]
vpc_id = "{{.nuon.install.inputs.vpc_id}}"
`

const inputsConfig = `#:schema https://api.nuon.co/v1/general/config-schema?source=inputs
[[group]]
name = "sandbox"
description = "Sandbox inputs"
display_name = "Sandbox inputs"

[[input]]
name = "vpc_id"
description = "vpc_id to install application into"
default = ""
sensitive = false
display_name = "VPC ID"
group = "sandbox"

[[input]]
name = "api_key"
description = "API key"
default = ""
sensitive = true
display_name = "API Key"
group = "sandbox"
`

const terraformComponentConfig = `#:schema https://api.nuon.co/v1/general/config-schema?source=terraform
name = "toml_terraform"
type = "terraform_module"
terraform_version = "1.5.3"

[connected_repo]
directory = "infra"
repo = "powertoolsdev/mono"
branch = "main"

[vars]
AWS_REGION = "{{.nuon.install.sandbox.account.region}}"
ACCOUNT_ID = "{{.nuon.install.sandbox.account.id}}"
`

const terraformInfraComponentConfig = `#:schema https://api.nuon.co/v1/general/config-schema?source=terraform
name = "toml_infra"
type = "terraform_module"
terraform_version = "1.5.4"

[connected_repo]
directory = "deployment"
repo = "powertoolsdev/mono"
branch = "main"

[vars]
iam_role = "{{.nuon.components.infra.outputs.iam_role}}"
`

const helmComponentConfig = `#:schema https://api.nuon.co/v1/general/config-schema?source=helm
name = "toml_helm"
type = "helm_chart"
chart_name = "e2e-helm"

[connected_repo]
directory = "deployment"
repo = "powertoolsdev/mono"
branch = "main"

[[values_file]]
contents = """
image.tag = {{.nuon.components.toml_docker_build.image.name}}
"""

[values]
"api.ingresses.public_domain" = "{{.nuon.components.infra.outputs.iam_role}}"
`

const dockerBuildComponentConfig = `#:schema https://api.nuon.co/v1/general/config-schema?source=docker_build
name = "toml_docker_build"
type = "docker_build"

dockerfile = "Dockerfile"

[connected_repo]
directory = "deployment"
repo = "powertoolsdev/mono"
branch = "main"
`

const containerImageComponentConfig = `#:schema https://api.nuon.co/v1/general/config-schema?source=container_image
name = "toml_container_image"
type = "container_image"

[public]
image_url = "kennethreitz/httpbin"
tag = "latest"
`

const jobComponentConfig = `#:schema https://api.nuon.co/v1/general/config-schema?source=job
name = "toml_job"
type = "job"

image_url = "{{.nuon.components.e2e_docker_build.image.repository.uri}}"
tag	  = "{{.nuon.components.e2e_docker_build.image.tag}}"
cmd	  = ["printenv"]
args	  = [""]

[env_vars]
PUBLIC_DOMAIN = "{{.nuon.components.infra.outputs.iam_role}}"
`

const ecrContainerImageComponentConfig = `#:schema https://api.nuon.co/v1/general/config-schema?source=container_image
name = "toml_container_image_ecr"
type = "container_image"

[aws_ecr]
iam_role_arn = "iam_role_arn"
image_url = "ecr-url"
tag = "latest"
region = "us-west-2"
`

