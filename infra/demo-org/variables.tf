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
