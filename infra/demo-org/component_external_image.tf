// run any image in ECR in a customer's cloud account
resource "nuon_container_image_component" "ecr_external" {
  name = "ecr_image"
  app_id = nuon_app.main.id

  public = {
    image_url = "kennethreitz/httpbin"
    tag = "latest"
  }

  // add a single build env var
  env_var {
    name = "MY_ENV_VAR"
    value = "{{.nuon.secrets.env_var}}"
  }
}

// sync any container image into your customer's cloud account
resource "nuon_container_image_component" "public_sync_only" {
  name = "docker_hub_sync_only"
  app_id = nuon_app.main.id

  public = {
    image_url = "kennethreitz/httpbin"
    tag = "latest"
  }
}
