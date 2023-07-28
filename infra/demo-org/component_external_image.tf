// run any image in ECR in a customer's cloud account
resource "nuon_container_image_component" "ecr_external" {
  name = "tf-managed-ecr-external"
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

  // add a single env var
  env_var {
    name = "MY_ENV_VAR"
    value = "{{.nuon.secrets.env_var}}"
  }

  // dynamically set env vars from another source
  dynamic "env_var" {
    for_each = local.default_env_vars
    iterator = ev
    content {
      name = ev.key
      value = ev.value
    }
  }
}

// sync any container image into your customer's cloud account
resource "nuon_container_image_component" "public_sync_only" {
  name = "tf-managed-docker-hub"
  app_id = nuon_app.main.id

  public = {
    image_url = "kennethreitz/httpbin"
    tag = "latest"
  }

  sync_only = true
}
