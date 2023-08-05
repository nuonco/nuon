module "repo" {
  source = "./ecr"

  name = "${var.install_id}-e2e"
  tags = {}

  region = var.region
}
