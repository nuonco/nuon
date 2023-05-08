module "repo" {
  source = "../modules/ecr"

  name = var.ecr_repo_name
  tags = {
    artifact      = "mob"
  }

  region = local.aws_settings.region
}
