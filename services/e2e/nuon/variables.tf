variable "org_id" {
  description = "org ID to create this app in. NOTE: the org must be connected to the powertoolsdev github."
  type = string
}

variable "api_auth_token" {
  description = "auth token. Recommend setting this using TF_VAR_api_auth_token"
  type = string
}

variable "app_name" {
  description = "App name, which can be useful when creating more than one instance of e2e in a single org."
  default = "e2e"
  type = string
}

# NOTE: eventually we will create these per run, so this can be self-containing.
variable "install_role_arn" {
  description = "IAM role ARN"
  type = string
}

variable "east_1_count" {
  description = "Number of installs to create in us-east-1"
  type = number
}

variable "east_2_count" {
  description = "Number of installs to create in us-east-2"
  type = number
}

variable "west_2_count" {
  description = "Number of installs to create in us-west-2"
  type = number
}
