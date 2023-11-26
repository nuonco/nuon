data "nuon_builtin_sandbox" "main" {
  name = "aws-eks"
}

resource "nuon_app_sandbox" "main" {
  app_id = nuon_app.main.id
  builtin_sandbox_release_id = data.nuon_builtin_sandbox.main.sandbox_release.id
  terraform_version = "v1.6.3"

  connected_repo = {
    repo = "powertoolsdev/mono"
    branch = "main"
    directory = "sandboxes/aws-eks"
  }

  input {
    name = "eks_version"
    value = "v1.27.8"
  }
}
