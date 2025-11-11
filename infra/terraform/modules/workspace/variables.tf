variable "auto_apply" {
  type    = bool
  default = false
}

variable "dir" {
  description = "terraform directory relative to the repo root"
  default     = ""
}

variable "name" {
  description = "workspace name"
  type        = string
}

variable "repo" {
  type        = string
  default     = ""
  description = <<EOT
  The GitHub repo to use as the source of this workspace.
  If not specified, no VCS connection will be made and the workspace can be applied locally.
  This should be the exception and never the norm.
  EOT
}

variable "slack_notifications_webhook_url" {
  description = "slack notifications webhook url for alerts"
  type        = string
  default     = ""
}

variable "pagerduty_service_account_id" {
  description = "Service account for creating Pagerduty incidents."
  type        = string
  default     = ""
}

variable "vars" {
  description = "variables to set on the workspace"
  type        = map(any)
  default     = {}
}

variable "env_vars" {
  description = "env variables to set on the workspace"
  type        = map(any)
  default     = {}
}

variable "variable_sets" {
  description = "names of variable sets to attach"
  type        = list(string)
  default     = []
}

variable "workspaces" {
  type    = list(any)
  default = ["stage", "prod"]
}

variable "project_id" {
  type = string
}

variable "terraform_version" {
  type    = string
  default = "1.7.5"
}

variable "trigger_workspaces" {
  type        = list(string)
  description = "workspace ids that should trigger runs of this workspace"
  default     = []
}

variable "trigger_prefixes" {
  type        = list(string)
  description = "additional prefixes that will trigger runs"
  default     = []
}

variable "vcs_branch" {
  type        = string
  description = "VCS branch to track for this workspace"
  default     = "main"
}
