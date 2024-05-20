variable "archived" {
  default     = false
  type        = bool
  description = "Whether to archive the repo or not"
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

variable "enable_ecr" {
  default     = false
  type        = bool
  description = "Whether to create an ECR repo for the source code repository"
}

variable "enable_stage_environment" {
  default     = false
  type        = bool
  description = "Whether to create a stage environment"
}

variable "enable_prod_environment" {
  default     = false
  type        = bool
  description = "Whether to create a prod environment"
}

variable "extra_ecr_repos" {
  description = "Extra repos to create, each of which will get a prefix of the name"
  type        = list(string)
  default     = []
}

variable "name" {
  type        = string
  description = "The repository name"
}

variable "owning_team_id" {
  description = "The owning team of the repo"
  type        = number
  default     = 4455826
}

variable "owning_team_name" {
  description = "The owning team name of the repo"
  type        = string
  default     = "powertoolsdev/team"
}


variable "prod_wait_timer" {
  type        = number
  default     = 15
  description = "Number of minutes to delay jobs for this environment"
}

variable "required_checks" {
  default     = ["Required PR Checks", "Required CI Checks"]
  type        = list(string)
  description = "Required checks that are enforced before merging"
}

variable "topics" {
  default     = []
  type        = list(string)
  description = "the list of topics to assign to the repo"
}

variable "is_public" {
  default     = false
  type        = bool
  description = "whether the repo should be public or not"
}

variable "is_private" {
  default     = false
  type        = bool
  description = "whether the repo should be private"
}

variable "is_fork" {
  default     = false
  type        = bool
  description = "whether the repo is a fork or not"
}

variable "collaborators" {
  type    = map(string)
  default = {}
}
