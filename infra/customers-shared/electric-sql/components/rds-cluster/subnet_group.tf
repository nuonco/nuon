module "subnet_group" {
  source  = "terraform-aws-modules/rds/aws//modules/db_subnet_group"
  version = "~> 6.0"

  name        = var.identifier
  description = "Subnet group for ${var.identifier}"
  subnet_ids  = local.subnet_ids
}
