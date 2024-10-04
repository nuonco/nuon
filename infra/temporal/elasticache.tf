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
