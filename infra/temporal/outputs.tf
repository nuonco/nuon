################################################################################
# Primary
################################################################################

output "enhanced_monitoring_iam_role_name" {
  description = "The name of the monitoring role"
  value       = module.primary.enhanced_monitoring_iam_role_name
}

output "enhanced_monitoring_iam_role_arn" {
  description = "The Amazon Resource Name (ARN) specifying the monitoring role"
  value       = module.primary.enhanced_monitoring_iam_role_arn
}

output "db_instance_address" {
  description = "The address of the RDS instance"
  value       = module.primary.db_instance_address
}

output "db_instance_arn" {
  description = "The ARN of the RDS instance"
  value       = module.primary.db_instance_arn
}

output "db_instance_availability_zone" {
  description = "The availability zone of the RDS instance"
  value       = module.primary.db_instance_availability_zone
}

output "db_instance_endpoint" {
  description = "The connection endpoint"
  value       = module.primary.db_instance_endpoint
}

output "db_instance_hosted_zone_id" {
  description = "The canonical hosted zone ID of the DB instance (to be used in a Route 53 Alias record)"
  value       = module.primary.db_instance_hosted_zone_id
}

output "db_instance_status" {
  description = "The RDS instance status"
  value       = module.primary.db_instance_status
}

output "db_instance_name" {
  description = "The database name"
  value       = module.primary.db_instance_name
}

output "db_instance_username" {
  description = "The master username for the database"
  value       = module.primary.db_instance_username
  sensitive   = true
}

output "db_instance_password" {
  description = "The master master for the database"
  value       = local.db_password
  sensitive   = true
}

output "db_instance_domain" {
  description = "The ID of the Directory Service Active Directory domain the instance is joined to"
  value       = module.primary.db_instance_domain
}

output "db_instance_domain_iam_role_name" {
  description = "The name of the IAM role to be used when making API calls to the Directory Service. "
  value       = module.primary.db_instance_domain_iam_role_name
}

output "db_instance_port" {
  description = "The database port"
  value       = module.primary.db_instance_port
}

output "db_instance_ca_cert_identifier" {
  description = "Specifies the identifier of the CA certificate for the DB instance"
  value       = module.primary.db_instance_ca_cert_identifier
}

output "db_subnet_group_id" {
  description = "The db subnet group name"
  value       = module.primary.db_subnet_group_id
}

output "db_subnet_group_arn" {
  description = "The ARN of the db subnet group"
  value       = module.primary.db_subnet_group_arn
}

output "db_parameter_group_id" {
  description = "The db parameter group id"
  value       = module.primary.db_parameter_group_id
}

output "db_parameter_group_arn" {
  description = "The ARN of the db parameter group"
  value       = module.primary.db_parameter_group_arn
}

# DB option group
output "db_option_group_id" {
  description = "The db option group id"
  value       = module.primary.db_option_group_id
}

output "db_option_group_arn" {
  description = "The ARN of the db option group"
  value       = module.primary.db_option_group_arn
}

################################################################################
# Replica
################################################################################

output "replica_enhanced_monitoring_iam_role_name" {
  description = "The name of the monitoring role"
  value       = try(module.replica[0].enhanced_monitoring_iam_role_name, "")
}

output "replica_enhanced_monitoring_iam_role_arn" {
  description = "The Amazon Resource Name (ARN) specifying the monitoring role"
  value       = try(module.replica[0].enhanced_monitoring_iam_role_arn, "")
}

output "replica_db_instance_address" {
  description = "The address of the RDS instance"
  value       = try(module.replica[0].db_instance_address, "")
}

output "replica_db_instance_arn" {
  description = "The ARN of the RDS instance"
  value       = try(module.replica[0].db_instance_arn, "")
}

output "replica_db_instance_availability_zone" {
  description = "The availability zone of the RDS instance"
  value       = try(module.replica[0].db_instance_availability_zone, "")
}

output "replica_db_instance_endpoint" {
  description = "The connection endpoint"
  value       = try(module.replica[0].db_instance_endpoint, "")
}

output "replica_db_instance_hosted_zone_id" {
  description = "The canonical hosted zone ID of the DB instance (to be used in a Route 53 Alias record)"
  value       = try(module.replica[0].db_instance_hosted_zone_id, "")
}

output "replica_db_instance_id" {
  description = "The RDS instance ID"
  value       = try(module.replica[0].db_instance_id, "")
}

output "replica_db_instance_resource_id" {
  description = "The RDS Resource ID of this instance"
  value       = try(module.replica[0].db_instance_resource_id, "")
}

output "replica_db_instance_status" {
  description = "The RDS instance status"
  value       = try(module.replica[0].db_instance_status, "")
}

output "replica_db_instance_name" {
  description = "The database name"
  value       = try(module.replica[0].db_instance_name, "")
}

output "replica_db_instance_username" {
  description = "The master username for the database"
  value       = try(module.replica[0].db_instance_username, "")
  sensitive   = true
}

output "replica_db_instance_domain" {
  description = "The ID of the Directory Service Active Directory domain the instance is joined to"
  value       = try(module.replica[0].db_instance_domain, "")
}

output "replica_db_instance_domain_iam_role_name" {
  description = "The name of the IAM role to be used when making API calls to the Directory Service. "
  value       = try(module.replica[0].db_instance_domain_iam_role_name, "")
}

output "replica_db_instance_port" {
  description = "The database port"
  value       = try(module.replica[0].db_instance_port, "")
}

output "replica_db_instance_ca_cert_identifier" {
  description = "Specifies the identifier of the CA certificate for the DB instance"
  value       = try(module.replica[0].db_instance_ca_cert_identifier, "")
}

output "replica_db_subnet_group_id" {
  description = "The db subnet group name"
  value       = try(module.replica[0].db_subnet_group_id, "")
}

output "replica_db_subnet_group_arn" {
  description = "The ARN of the db subnet group"
  value       = try(module.replica[0].db_subnet_group_arn, "")
}

output "replica_db_parameter_group_id" {
  description = "The db parameter group id"
  value       = try(module.replica[0].db_parameter_group_id, "")
}

output "replica_db_parameter_group_arn" {
  description = "The ARN of the db parameter group"
  value       = try(module.replica[0].db_parameter_group_arn, "")
}

# DB option group
output "replica_db_option_group_id" {
  description = "The db option group id"
  value       = try(module.replica[0].db_option_group_id, "")
}

output "replica_db_option_group_arn" {
  description = "The ARN of the db option group"
  value       = try(module.replica[0].db_option_group_arn, "")
}

################################################################################
# CloudWatch Log Groups
################################################################################

output "db_instance_cloudwatch_log_groups_primary" {
  description = "Map of CloudWatch log groups created and their attributes"
  value       = module.primary.db_instance_cloudwatch_log_groups
}

output "replica_db_instance_cloudwatch_log_groups_primary" {
  description = "Map of CloudWatch log groups created and their attributes"
  value       = try(module.replica[0].db_instance_cloudwatch_log_groups, null)
}

################################################################################
# DNS Entries
################################################################################

output "primary_friendly_dns_name" {
  description = "The hostname of the ALIAS in the private hosted zone of the primary"
  value       = aws_route53_record.primary.fqdn
}

output "replica_friendly_dns_name" {
  description = "The hostname of the ALIAS in the private hosted zone of the replica"
  value       = try(aws_route53_record.replica[0].fqdn, "")
}

output "elasticache_friendly_dns_name" {
  description = "The hostname of the friendly DNS name in the private hosted zone"
  value       = try(aws_route53_record.elasticache[0].fqdn, "")
}

output "elasticsearch_friendly_dns_name" {
  description = "The hostname of the friendly DNS name in the private hosted zone"
  value       = try(aws_route53_record.elasticsearch[0].fqdn, "")
}

################################################################################
# Helm Release
################################################################################

# NOTE(jdt): this is always an apply behind :sob:
# output "manifests" {
#   value = helm_release.temporal.manifest
# }

output "frontend_url" {
  value = local.temporal.frontend_url
}

output "web_url" {
  value = local.temporal.web_url
}

output "image_tag" {
  value = local.temporal.image_tag
}

output "helm_version" {
  value = local.temporal.version
}

