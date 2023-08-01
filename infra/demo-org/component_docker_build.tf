// build any docker image from source and run it in a customer's cloud account
resource "nuon_docker_build_component" "public_docker" {
  name = "Public Repo Docker Build"
  app_id = nuon_app.main.id

  dockerfile = "Dockerfile"

  public_repo = {
    directory = "."
    repo = "https://github.com/jonmorehouse/go-httpbin.git"
    branch = "main"
  }

  basic_deploy = {
    port = 8080
    instance_count = 1
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

// build a docker image and sync it into your customer's cloud account
resource "nuon_docker_build_component" "private_docker" {
  name = "Private Repo Docker Build (sync only)"
  app_id = nuon_app.main.id

  dockerfile = "Dockerfile"
  connected_repo = {
    directory = "demo/components/go-httpbin"
    repo = "powertoolsdev/demo"
    branch = "main"
  }

  sync_only = true
}
