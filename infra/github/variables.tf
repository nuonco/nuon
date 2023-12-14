data "utils_deep_merge_yaml" "vars" {
  input = [
    file("vars/defaults.yaml"),
  ]
}

data "utils_deep_merge_yaml" "mono_vars" {
  input = [
    file("vars/mono.yaml"),
  ]
}

locals {
  vars      = yamldecode(data.utils_deep_merge_yaml.vars.output)
  mono_vars = yamldecode(data.utils_deep_merge_yaml.mono_vars.output)
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

// nuonco_shared vars
variable "nuonco_shared_app_id" {
  type = string
}

variable "nuonco_shared_app_installation_id" {
  type = string
}

variable "nuonco_shared_app_pem_file" {
  type      = string
  sensitive = true
}
