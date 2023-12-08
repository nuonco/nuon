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
  default     = "nuonco/sandboxes"
}

variable "sandbox_dir" {
  description = "Sandbox directory to use."
}

variable "sandbox_branch" {
  description = "Sandbox branch to use."
  default     = "main"
}

variable "install_inputs" {
  type = list(object({
    name          = string
    description   = string
    default       = string
    required      = bool
    value         = string
    interpolation = string
  }))
  description = "Inputs that will be interpolated per install."

  # The previously hard-coded install inputs. We'll need to update all callers of this module to remove this.
  default = [
    {
      name          = "eks_version"
      description   = "Version of k8s to use with EKS."
      default       = ""
      required      = true
      value         = "v1.27.8"
      interpolation = "{{.nuon.install.inputs.eks_version}}"
    },
    {
      name          = "admin_access_role_arn"
      description   = "The IAM role that provides access to manage the install."
      default       = "default"
      required      = false
      value         = "arn:aws:iam::676549690856:role/aws-reserved/sso.amazonaws.com/us-east-2/AWSReservedSSO_NuonAdmin_b8aea3365312317b"
      interpolation = "{{.nuon.install.inputs.admin_access_role_arn"
    },
  ]
}
