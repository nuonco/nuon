package service

// Deprecated
const flatAppConfigTemplate = `# This file contains template values for common Nuon application configuration options.
# To use it for your app, edit as needed, then rename this file to %s and run
#
#   nuon apps sync -c %s
#
# See https://docs.nuon.co/concepts/apps for more information.

version = "v1"

[installer]
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
footer_markdown = ""

[runner]
runner_type = "aws-ecs"

[[runner.env_var]]
name = "runner-env-var"
value = "runner-env-var"

[sandbox]
terraform_version = "1.5.4"

# https://docs.nuon.co/guides/install-access-delegation#setup-delegation
# if you are using delegation, otherwise remove
# govcloud clients must reach out for additional configuration
aws_delegation_iam_role_arn = "arn:aws:iam::xxxxxxxxxxxx:role/nuon-aws-ecs-install-access"

[sandbox.public_repo]
directory = "aws-ecs-byo-vpc"
repo = "nuonco/sandboxes"
branch = "main"

[[sandbox.var]]
name = "vpc_id"
value = "{{.nuon.install.inputs.vpc_id}}"

[inputs]
[[inputs.group]]
name = "sandbox"
description = "Sandbox inputs"
display_name = "Sandbox inputs"

[[inputs.input]]
name = "vpc_id"
description = "vpc_id to install application into"
default = ""
sensitive = false
display_name = "VPC ID"
group = "sandbox"

[[inputs.input]]
name = "api_key"
description = "API key"
default = ""
sensitive = true
display_name = "API Key"
group = "sandbox"

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

[[components.values_file]]
contents = """
image.tag = {{.nuon.components.toml_docker_build.image.name}}
"""

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
`

const topLevelConfig = `# This file contains template values for common Nuon application configuration options.
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
directory = "aws-ecs-byo-vpc"
repo = "nuonco/sandboxes"
branch = "main"

[vars]
vpc_id = "{{.nuon.install.inputs.vpc_id}}"
`

const inputsConfig = `#:schema https://api.nuon.co/v1/general/config-schema?source=input
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