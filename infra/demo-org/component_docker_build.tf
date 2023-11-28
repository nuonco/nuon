// build any docker image from source and run it in a customer's cloud account
resource "nuon_docker_build_component" "public_docker" {
  name = "public_repo_docker_build"
  app_id = nuon_app.main.id

  dockerfile = "Dockerfile"

  public_repo = {
    directory = "."
    repo = "https://github.com/jonmorehouse/go-httpbin.git"
    branch = "main"
  }

  sync_only = true
}

// build a docker image and sync it into your customer's cloud account
resource "nuon_docker_build_component" "private_docker" {
  name = "private_repo_docker_build"
  app_id = nuon_app.main.id

  dockerfile = "Dockerfile"
  connected_repo = {
    directory = "demo/components/go-httpbin"
    repo = "powertoolsdev/demo"
    branch = "main"
  }

  sync_only = true
}
