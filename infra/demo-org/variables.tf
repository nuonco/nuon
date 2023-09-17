// defining values to set across the board
locals {
  default_env_vars = {
    NUON_SECRET = "secret"
    NUON_CONNECTION = "connection"
    NUON_OUTPUTS = "outputs"
  }

  default_secrets = {
    NUON_SECRET = "secret"
    NUON_CONNECTION = "connection"
    NUON_OUTPUTS = "outputs"
  }
}

variable "api_url" {
  description = "api_url set by standard api variable set"
}

variable "api_token" {
  description = "api_token set by standard api variable set"
}

variable "org_id" {
  description = "org id set from vars"
}
