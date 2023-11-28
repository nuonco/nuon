// run any image in ECR in a customer's cloud account
resource "nuon_container_image_component" "ecr_external" {
  name = "ecr_image"
  app_id = nuon_app.main.id

  public = {
    image_url = "kennethreitz/httpbin"
    tag = "latest"
  }

  basic_deploy = {
    port = 8080
    instance_count = 5
    health_check_path = "/"

  }

  sync_only = true
}

// sync any container image into your customer's cloud account
resource "nuon_container_image_component" "public_sync_only" {
  name = "docker_hub_sync_only"
  app_id = nuon_app.main.id

  public = {
    image_url = "kennethreitz/httpbin"
    tag = "latest"
  }

  sync_only = true
}
