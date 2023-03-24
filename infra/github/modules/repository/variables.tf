variable "name" {
  type        = string
  description = "The repository name"
}

variable "description" {
  type        = string
  description = "The repository description"
}

variable "enable_branch_protection" {
  default     = true
  type        = bool
  description = "Enable branch protection. Disable with caution."
}

variable "topics" {
  default     = []
  type        = list(string)
  description = "the list of topics to assign to the repo"
}

variable "enable_ecr" {
  default     = false
  type        = bool
  description = "Whether to create an ECR repo for the source code repository"
}

variable "archived" {
  default     = false
  type        = bool
  description = "Whether to archive the repo or not"
}

variable "enable_prod_environment" {
  default     = false
  type        = bool
  description = "Whether to create a prod environment"
}

variable "prod_wait_timer" {
  type        = number
  default     = 15
  description = "Number of minutes to delay jobs for this environment"
}

variable "enable_stage_environment" {
  default     = false
  type        = bool
  description = "Whether to create a stage environment"
}

variable "owning_team" {
  description = "The owning team of the repo"
  type        = map(any)
  default     = {}
}

variable "extra_ecr_repos" {
  description = "Extra repos to create, each of which will get a prefix of the name"
  type        = list(string)
  default     = []
}

variable "is_template" {
  default     = false
  type        = bool
  description = "Whether the repo is a template repo"
}
