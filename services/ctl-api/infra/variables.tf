data "utils_deep_merge_yaml" "vars" {
  input = [
    file("vars/defaults.yaml"),
    file("vars/${var.env}.yaml"),
  ]
}

locals {
  name = "ctl-api"
  vars = yamldecode(data.utils_deep_merge_yaml.vars.output)
}

variable "env" {
  type        = string
  description = "env"
}
