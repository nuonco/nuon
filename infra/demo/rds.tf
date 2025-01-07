################################################################################
# Dependencies
################################################################################
module "security_group_rds" {
  source  = "terraform-aws-modules/security-group/aws"
  version = "~> 5.0"

  name        = local.rds.db_name
  description = "RDS security group for ${local.rds.db_name}."
  vpc_id      = module.byovpc.vpc.id

  # ingress
  ingress_with_cidr_blocks = [
    {
      from_port   = local.rds.port
      to_port     = local.rds.port
      protocol    = "tcp"
      description = "RDS access from within VPC"
      cidr_blocks = module.byovpc.vpc.private_subnet_cidr_blocks[0]
    },
  ]
}

################################################################################
# Master DB
################################################################################
module "primary" {
  source  = "terraform-aws-modules/rds/aws"
  version = "~> 5.0"

  identifier = "primary-${local.rds.db_name}"

  allow_major_version_upgrade  = true
  engine                       = local.rds.engine
  engine_version               = local.rds.engine_version
  family                       = local.rds.family
  major_engine_version         = local.rds.major_engine_version
  instance_class               = local.rds.instance_class
  performance_insights_enabled = true

  allocated_storage = local.rds.allocated_storage

  db_name  = local.rds.db_name
  username = local.rds.username
  port     = local.rds.port

  multi_az               = local.rds.multi_az
  db_subnet_group_name   = module.byovpc.vpc.db_subnet_group_id
  vpc_security_group_ids = [module.security_group_rds.security_group_id]

  maintenance_window = "Mon:00:00-Mon:03:00"
  backup_window      = "03:00-06:00"

  # Backups are required in order to create a replica
  backup_retention_period = local.rds.backup_retention_period
  skip_final_snapshot     = local.rds.skip_final_snapshot
  deletion_protection     = local.rds.deletion_protection
  storage_encrypted       = local.rds.storage_encrypted

  iam_database_authentication_enabled = true
  apply_immediately                   = true
}
