variable "app_name" {
  description = "App name, which can be useful when creating more than one instance of e2e in a single org."
  default     = "e2e"
  type        = string
}

variable "create_components" {
  description = "Control whether components will be created or not"
  default     = true
  type        = bool
}

variable "component_prefix" {
  description = "Prefix to add onto each component, to make looking up by name easier"
  default     = ""
  type        = string
}

variable "app_runner_type" {
  description = "app runner type"
  default     = "aws-eks"
  type        = string
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

variable "install_count" {
  description = "install count"
  default     = 0
  type        = number
}

variable "install_prefix" {
  default = "e2e-"
  type    = string
}

variable "aws" {
  type = list(object({
    regions      = list(string)
    iam_role_arn = string
  }))
  description = "Inputs for an aws e2e install"
  default     = []
}

variable "azure" {
  type = list(object({
    locations                  = list(string)
    subscription_id            = string
    subscription_tenant_id     = string
    service_principal_app_id   = string
    service_principal_password = string
  }))
  description = "Inputs for an azure e2e install"
  default     = []
}

variable "groups" {
  type = list(object({
    name         = string
    description  = string
    display_name = string
  }))
  description = "Input groups"

  default = [
    {
      name         = "sandbox_inputs"
      display_name = "Sandbox Inputs"
      description  = "Inputs related to the sandbox"
    },
    {
      name         = "app_inputs"
      display_name = "App Inputs"
      description  = "Inputs related to the app"
    },
  ]
}

variable "inputs" {
  type = list(object({
    name          = string
    description   = string
    default       = string
    required      = bool
    value         = string
    group         = string
    interpolation = string
    display_name  = string
    sensitive     = bool
  }))
  description = "Inputs that will be used for app inputs, and then set on each install"

  default = [
    {
      name          = "cluster_version"
      display_name  = "Cluster Version"
      description   = "Version of k8s to use for EKS."
      default       = ""
      group         = "sandbox_inputs"
      required      = true
      value         = "v1.27.8"
      interpolation = "{{.nuon.install.inputs.cluster_version}}"
      sensitive     = false
    },
    {
      name          = "admin_access_role_arn"
      display_name  = "Admin Access Role ARN"
      group         = "sandbox_inputs"
      description   = "The IAM role that provides access to manage the install."
      default       = "default"
      required      = false
      value         = "arn:aws:iam::676549690856:role/aws-reserved/sso.amazonaws.com/us-east-2/AWSReservedSSO_NuonAdmin_b8aea3365312317b"
      interpolation = "{{.nuon.install.inputs.admin_access_role_arn}}"
      sensitive     = false
    },
    {
      name          = "api_key"
      display_name  = "API Key"
      description   = "API key to access a third party provider"
      default       = ""
      group         = "app_inputs"
      required      = true
      value         = "D066077E-F464-47F1-90EE-FE2466D0561C"
      interpolation = "{{.nuon.install.inputs.api_key"
      sensitive     = true
    },
  ]
}
