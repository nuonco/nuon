module "dev_public_ecr" {
  source = "../modules/public-ecr"

  name        = "dev-public"
  region      = local.aws_settings.public_region
  description = "ECR repo for development"
  about       = "ECR repo for pushing development containers that need to be public, such as for testing plugins."
  tags = {}

  providers = {
    aws = aws.public
  }
}
