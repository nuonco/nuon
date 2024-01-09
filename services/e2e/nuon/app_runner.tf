resource "nuon_app_runner" "main" {
  app_id = nuon_app.main.id

  runner_type = "aws-eks"
  env_var {
    name = "NUON_RUNNER_TYPE"
    value = "aws-eks"
  }
}
