resource "nuon_app_runner" "main" {
  app_id = nuon_app.main.id

  runner_type = var.app_runner_type
}
