resource "nuon_job_component" "e2e" {
  name      = "e2e_job"
  app_id    = nuon_app.main.id
  image_url = "bitnami/kubectl"
  tag       = "latest"
  cmd       = ["kubectl"]
  args      = ["version"]
}
