################################################################################
# Dependencies
################################################################################
module "security_group_rds" {
  source  = "terraform-aws-modules/security-group/aws"
  version = "~> 5.0"

  name        = local.name
  description = "RDS security group for ${local.name} in ${var.env}"
  vpc_id      = data.aws_vpc.vpc.id

  # ingress
  ingress_with_cidr_blocks = [
    {
      from_port   = local.vars.rds.port
      to_port     = local.vars.rds.port
      protocol    = "tcp"
      description = "RDS access from within VPC"
      cidr_blocks = data.aws_vpc.vpc.cidr_block_associations[0].cidr_block
    },
  ]
}

module "subnet_group" {
  source  = "terraform-aws-modules/rds/aws//modules/db_subnet_group"
  version = "~> 6.0"

  name        = local.name
  description = "Subnet group for ${local.name} in ${var.env}"
  subnet_ids  = data.aws_subnets.private.ids
}

################################################################################
# Master DB
################################################################################
module "primary" {
  source  = "terraform-aws-modules/rds/aws"
  version = "~> 5.0"

  identifier = "primary-${local.name}"

  engine               = local.vars.rds.engine
  engine_version       = local.vars.rds.engine_version
  family               = local.vars.rds.family
  parameters           = local.vars.rds.parameters
  major_engine_version = local.vars.rds.major_engine_version
  instance_class       = local.vars.rds.instance_class

  allocated_storage = local.vars.rds.allocated_storage

  db_name  = local.vars.rds.db_name
  username = local.vars.rds.username
  port     = local.vars.rds.port

  multi_az               = local.vars.rds.multi_az
  db_subnet_group_name   = module.subnet_group.db_subnet_group_id
  vpc_security_group_ids = [module.security_group_rds.security_group_id]

  maintenance_window              = "Mon:00:00-Mon:03:00"
  backup_window                   = "03:00-06:00"
  enabled_cloudwatch_logs_exports = local.vars.rds.enabled_cloudwatch_logs_exports

  # Backups are required in order to create a replica
  backup_retention_period = local.vars.rds.backup_retention_period
  skip_final_snapshot     = local.vars.rds.skip_final_snapshot
  deletion_protection     = local.vars.rds.deletion_protection
  storage_encrypted       = local.vars.rds.storage_encrypted

  iam_database_authentication_enabled = true
  apply_immediately                   = true
}

resource "aws_route53_record" "primary" {
  zone_id = data.aws_route53_zone.private.id
  name    = "${local.name}-primary.db.${data.aws_route53_zone.private.name}"
  type    = "A"

  alias {
    name                   = module.primary.db_instance_address
    zone_id                = module.primary.db_instance_hosted_zone_id
    evaluate_target_health = false
  }
}
