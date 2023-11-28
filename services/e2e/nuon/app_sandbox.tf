data "nuon_builtin_sandbox" "main" {
  name = "aws-eks"
}

resource "nuon_app_sandbox" "main" {
  app_id = nuon_app.main.id
  terraform_version = "v1.6.3"

  public_repo = {
    repo = "nuonco/sandboxes"
    branch = "main"
    directory = "aws-eks"
  }

  input {
    name = "eks_version"
    value = "v1.27.8"
  }
}
