data "nuon_builtin_sandbox" "main" {
  name = "aws-eks"
}

resource "nuon_app_sandbox" "main" {
  app_id            = nuon_app.main.id
  terraform_version = "v1.6.3"

  public_repo = {
    repo      = var.sandbox_repo
    branch    = var.sandbox_branch
    directory = var.sandbox_dir
  }

  var {
    name  = "eks_version"
    value = "v1.27.8"
  }

  var {
    name  = "admin_access_role_arn"
    value = "arn:aws:iam::676549690856:role/aws-reserved/sso.amazonaws.com/us-east-2/AWSReservedSSO_NuonAdmin_b8aea3365312317b"
  }
}
