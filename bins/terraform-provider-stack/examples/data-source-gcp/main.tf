# Example: an install-stacks/gcp module reading its configuration from the Nuon
# control plane via the stack_config data source, instead of receiving it as
# generated tfvars. The customer supplies only the phone_home_id (plus the GCP
# project/region, which are not known server-side).

terraform {
  required_providers {
    stack = {
      source = "nuonco/stack"
    }
  }
}

provider "stack" {
  # Defaults to https://api.nuon.co
  # api_url = "https://api.nuon.co"
}

variable "phone_home_id" {
  type = string
}

variable "gcp_project_id" {
  type = string
}

variable "gcp_region" {
  type = string
}

data "stack_config" "this" {
  phone_home_id = var.phone_home_id
}

module "stack" {
  source = "git::https://github.com/nuonco/install-stacks.git//gcp"

  # Customer-supplied (not known server-side).
  gcp_project_id = var.gcp_project_id
  gcp_region     = var.gcp_region

  # Read from the control plane instead of tfvars.
  nuon_install_id = data.stack_config.this.install_id
  nuon_org_id     = data.stack_config.this.org_id
  nuon_app_id     = data.stack_config.this.app_id

  runner_id              = data.stack_config.this.runner_id
  runner_api_url         = data.stack_config.this.runner_api_url
  runner_api_token       = data.stack_config.this.gcp.runner_api_token
  runner_init_script_url = data.stack_config.this.gcp.runner_init_script_url
  phone_home_url         = data.stack_config.this.phone_home_url

  provision_permissions   = data.stack_config.this.gcp.provision_permissions
  maintenance_permissions = data.stack_config.this.gcp.maintenance_permissions
  deprovision_permissions = data.stack_config.this.gcp.deprovision_permissions

  provision_predefined_role   = data.stack_config.this.gcp.provision_predefined_role
  maintenance_predefined_role = data.stack_config.this.gcp.maintenance_predefined_role
  deprovision_predefined_role = data.stack_config.this.gcp.deprovision_predefined_role

  break_glass_roles = data.stack_config.this.gcp.break_glass_roles
  custom_roles      = data.stack_config.this.gcp.custom_roles

  install_inputs        = data.stack_config.this.install_inputs
  auto_generate_secrets = data.stack_config.this.auto_generate_secrets
  secrets               = data.stack_config.this.secrets
}
