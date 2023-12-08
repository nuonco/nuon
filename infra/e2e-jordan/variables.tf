locals {
  name             = "e2e-jordan"
  sandboxes_repo   = "nuonco/sandboxes"
  sandboxes_branch = "main"
}

variable "org_id" {
  description = "org id"
  default     = "orgzblonf9hol7jq92vkdriio4"
}

variable "sandbox_org_id" {
  description = "sandbox org id"
  default     = "orgvwpbd584d7v7o9x8oxqfo6b"
}
