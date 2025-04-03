module "replica" {
  count   = local.vars.rds.enable_replica ? 1 : 0
  source  = "terraform-aws-modules/rds/aws"

  identifier = "${local.vars.rds.db_name}-replica"

  replicate_source_db    = module.primary.db_instance_arn

  engine               = local.vars.rds.engine
  engine_version       = local.vars.rds.engine_version
  family               = local.vars.rds.family
  allow_major_version_upgrade = true
  major_engine_version = local.vars.rds.major_engine_version
  instance_class       = local.vars.rds.instance_class

  allocated_storage = local.vars.rds.allocated_storage

  multi_az               = local.vars.rds.multi_az
  vpc_security_group_ids = [module.security_group_rds.security_group_id]

  maintenance_window              = "Wed:00:00-Wed:03:00"
  backup_window                   = "03:00-06:00"
  enabled_cloudwatch_logs_exports = local.vars.rds.enabled_cloudwatch_logs_exports

  backup_retention_period = 0
  skip_final_snapshot     = true
  deletion_protection     = false
  storage_encrypted       = local.vars.rds.storage_encrypted

  iam_database_authentication_enabled = true
  apply_immediately                   = true
}

resource "aws_route53_record" "replica" {
  count   = local.vars.rds.enable_replica ? 1 : 0
  zone_id = data.aws_route53_zone.private.id
  name    = "${local.name}-replica.db.${data.aws_route53_zone.private.name}"
  type    = "A"

  alias {
    name                   = module.replica[0].db_instance_address
    zone_id                = module.replica[0].db_instance_hosted_zone_id
    evaluate_target_health = false
  }
}
