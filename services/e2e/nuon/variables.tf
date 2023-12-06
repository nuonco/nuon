variable "app_name" {
  description = "App name, which can be useful when creating more than one instance of e2e in a single org."
  default     = "e2e"
  type        = string
}

variable "install_role_arn" {
  description = "IAM role ARN"
  type        = string
}

variable "east_1_count" {
  description = "Number of installs to create in us-east-1"
  type        = number
}

variable "east_2_count" {
  description = "Number of installs to create in us-east-2"
  type        = number
}

variable "west_2_count" {
  description = "Number of installs to create in us-west-2"
  type        = number
}

variable "sandbox_repo" {
  description = "Sandbox repository to use, must be public."
  default = "nuonco/sandboxes"
}

variable "sandbox_dir" {
  description = "Sandbox directory to use."
}

variable "sandbox_branch" {
  description = "Sandbox branch to use."
  default = "main"
}
