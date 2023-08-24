data "utils_deep_merge_yaml" "vars" {
  input = [
    file("vars/defaults.yaml"),
  ]
}

locals {
  vars = yamldecode(data.utils_deep_merge_yaml.vars.output)
}

// powertoolsdev vars
variable "powertools_app_id" {
  type = string
}

variable "powertools_app_installation_id" {
  type = string
}

variable "powertools_app_pem_file" {
  type      = string
  sensitive = true
}

// nuonco vars
variable "nuonco_app_id" {
  type = string
}

variable "nuonco_install_id" {
  type = string
}

variable "nuonco_pem_file" {
  type      = string
  sensitive = true
}
