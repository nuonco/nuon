module "repo" {
  source = "./ecr"

  name = var.ecr_repo_name
  tags = {
    artifact = "demo-ecr"
  }

  region = local.aws_settings.region
}
