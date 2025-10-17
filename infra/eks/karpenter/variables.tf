variable "cluster_name" {}
variable "cluster_endpoint" {}

variable "namespace" {}

variable "karpenter_version" {}

variable "discovery_key" {}

variable "discovery_value" {}

variable "oidc_provider_arn" {}

variable "node_iam_role_arn" {}
variable "node_iam_role_name" {}

variable "tags" {}

variable "ec2nodeclasses" {
  description = "List of EC2NodeClasses to create"
  type = list(object({
    name = string
    block_devices = optional(object({
      device_name = string
      volume_size = string
      volume_type = string
    }))
  }))
  default = []
}
variable "instance_types" {}

variable "additional_nodepools" {}
