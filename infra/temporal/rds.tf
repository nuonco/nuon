locals {
  vpc = {
    id         = data.aws_vpc.vpc.id
    cidr_block = data.aws_vpc.vpc.cidr_block_associations[0].cidr_block
  }
}

################################################################################
# Dependencies
################################################################################
module "security_group_rds" {
  source  = "terraform-aws-modules/security-group/aws"

  name        = local.name
  description = "RDS security group for ${local.name}"
  vpc_id      = local.vpc.id

  # ingress
  ingress_with_cidr_blocks = [
    {
      from_port   = local.vars.rds.port
      to_port     = local.vars.rds.port
      protocol    = "tcp"
      description = "RDS access from within VPC"
      cidr_blocks = local.vpc.cidr_block
    },
  ]
}

module "subnet_group" {
  source  = "terraform-aws-modules/rds/aws//modules/db_subnet_group"

  name        = local.name
  description = "Subnet group for ${local.name}"
  subnet_ids  = data.aws_subnets.private.ids
}

################################################################################
# Master DB
################################################################################
module "primary" {
  source  = "terraform-aws-modules/rds/aws"

  identifier = "primary-${local.name}"

  engine               = local.vars.rds.engine
  engine_version       = local.vars.rds.engine_version
  family               = local.vars.rds.family
  major_engine_version = local.vars.rds.major_engine_version
  instance_class       = local.vars.rds.instance_class

  allocated_storage = local.vars.rds.allocated_storage

  db_name  = local.vars.rds.db_name
  username = local.vars.rds.username
  port     = local.vars.rds.port

  multi_az               = local.vars.rds.multi_az
  db_subnet_group_name   = module.subnet_group.db_subnet_group_id
  vpc_security_group_ids = [module.security_group_rds.security_group_id]
  manage_master_user_password = true
  manage_master_user_password_rotation = true
  master_user_password_rotation_automatically_after_days = 365

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
  name    = "${local.service}-primary.db.${data.aws_route53_zone.private.name}"
  type    = "A"

  alias {
    name                   = module.primary.db_instance_address
    zone_id                = module.primary.db_instance_hosted_zone_id
    evaluate_target_health = false
  }
}

################################################################################
# Replica DB
################################################################################

module "replica" {
  count   = local.vars.rds.enable_replica ? 1 : 0
  source  = "terraform-aws-modules/rds/aws"

  identifier = "replica-${local.name}"

  # Source database. For cross-region use db_instance_arn
  replicate_source_db    = module.primary.db_instance_identifier

  engine               = local.vars.rds.engine
  engine_version       = local.vars.rds.engine_version
  family               = local.vars.rds.family
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
  name    = "${local.service}-replica.db.${data.aws_route53_zone.private.name}"
  type    = "A"

  alias {
    name                   = module.replica[0].db_instance_address
    zone_id                = module.replica[0].db_instance_hosted_zone_id
    evaluate_target_health = false
  }
}

################################################################################
# Elasticache
################################################################################

module "security_group_elasticache" {
  count   = local.vars.elasticache.enabled ? 1 : 0
  source  = "terraform-aws-modules/security-group/aws"

  name        = local.name
  description = "elasticache security group for ${local.name}"
  vpc_id      = local.vpc.id

  # ingress
  ingress_with_cidr_blocks = [
    {
      from_port   = local.vars.elasticache.port
      to_port     = local.vars.elasticache.port
      protocol    = "tcp"
      description = "elasticache access from within VPC"
      cidr_blocks = local.vpc.cidr_block
    },
  ]
}

resource "aws_elasticache_subnet_group" "elasticache" {
  count      = local.vars.elasticache.enabled ? 1 : 0
  name       = "elasticache-${local.name}"
  subnet_ids = data.aws_subnets.private.ids
}

resource "aws_elasticache_cluster" "elasticache" {
  count                = local.vars.elasticache.enabled ? 1 : 0
  cluster_id           = "elasticache-${local.name}"
  engine               = local.vars.elasticache.engine
  node_type            = local.vars.elasticache.node_type
  num_cache_nodes      = local.vars.elasticache.num_cache_nodes
  parameter_group_name = aws_elasticache_parameter_group.elasticache[0].id
  engine_version       = local.vars.elasticache.engine_version
  port                 = local.vars.elasticache.port
  security_group_ids   = [module.security_group_elasticache[0].security_group_id]
  subnet_group_name    = aws_elasticache_subnet_group.elasticache[0].name
}

resource "aws_elasticache_parameter_group" "elasticache" {
  count  = local.vars.elasticache.enabled ? 1 : 0
  name   = "params-${local.name}"
  family = local.vars.elasticache.family

  dynamic "parameter" {
    for_each = local.vars.elasticache.parameters
    content {
      name  = parameter.key
      value = parameter.value
    }
  }
}

resource "aws_route53_record" "elasticache" {
  count   = local.vars.elasticache.enabled ? 1 : 0
  zone_id = data.aws_route53_zone.private.id
  name    = "${local.service}.cache.${data.aws_route53_zone.private.name}"
  type    = "CNAME"
  ttl     = "300"
  records = [aws_elasticache_cluster.elasticache[0].cache_nodes[0].address]
}

################################################################################
# Elasticsearch
################################################################################
locals {
  elasticsearch = {
    custom_domain = "${local.service}.search.${data.aws_route53_zone.private.name}"
  }
}

module "security_group_elasticsearch" {
  count   = local.vars.elasticsearch.enabled ? 1 : 0
  source  = "terraform-aws-modules/security-group/aws"

  name        = local.name
  description = "elasticsearch security group for ${local.name}"
  vpc_id      = local.vpc.id

  # ingress
  ingress_with_cidr_blocks = [
    {
      from_port   = 443
      to_port     = 443
      protocol    = "tcp"
      description = "elasticsearch access from within VPC"
      cidr_blocks = local.vpc.cidr_block
    },
  ]
}

resource "aws_opensearch_domain" "elasticsearch" {
  count          = local.vars.elasticsearch.enabled ? 1 : 0
  domain_name    = local.name
  engine_version = "OpenSearch_1.0"

  cluster_config {
    instance_type  = local.vars.elasticsearch.instance_type
    instance_count = local.vars.elasticsearch.instance_count

    zone_awareness_enabled = local.vars.elasticsearch.zone_awareness.enabled
    zone_awareness_config {
      availability_zone_count = local.vars.elasticsearch.zone_awareness.availability_zone_count
    }

    warm_enabled = local.vars.elasticsearch.warm.enabled
    warm_count   = local.vars.elasticsearch.warm.count

    dedicated_master_enabled = local.vars.elasticsearch.dedicated_master.enabled
    dedicated_master_type    = local.vars.elasticsearch.dedicated_master.type
    dedicated_master_count   = local.vars.elasticsearch.dedicated_master.count
  }

  ebs_options {
    ebs_enabled = local.vars.elasticsearch.ebs.enabled
    volume_size = local.vars.elasticsearch.ebs.volume_size
  }

  vpc_options {
    subnet_ids = slice(data.aws_subnets.private.ids, 0, local.vars.elasticsearch.instance_count)

    security_group_ids = [module.security_group_elasticsearch[0].security_group_id]
  }

  domain_endpoint_options {
    custom_endpoint_enabled = true
    custom_endpoint         = local.elasticsearch.custom_domain
    enforce_https           = false
    tls_security_policy     = "Policy-Min-TLS-1-2-2019-07"
  }

  encrypt_at_rest {
    enabled = true
  }

  node_to_node_encryption {
    enabled = true
  }

  advanced_options = local.vars.elasticsearch.advanced_options
}

resource "aws_route53_record" "elasticsearch" {
  count   = local.vars.elasticsearch.enabled ? 1 : 0
  zone_id = data.aws_route53_zone.private.id
  name    = local.elasticsearch.custom_domain
  type    = "CNAME"
  ttl     = "300"
  records = [aws_opensearch_domain.elasticsearch[0].endpoint]
}
