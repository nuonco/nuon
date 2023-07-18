output "db_instance_address" {
  value = aws_route53_record.primary.name
}

output "db_instance_host" {
  value = module.primary.db_instance_address
}

output "db_instance_admin_name" {
  value = module.primary.db_instance_name
}

output "db_instance_name" {
  value = "api"
}

output "db_instance_port" {
  value = module.primary.db_instance_port
}

output "db_instance_admin_username" {
  sensitive = true
  value     = module.primary.db_instance_username
}

output "db_instance_username" {
  sensitive = true
  value     = "api"
}

output "org_iam_role_name_templates" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.org_iam_role_name_templates)
}

output "buckets" {
  description = "buckets used to store vendor data"
  value       = nonsensitive(data.tfe_outputs.infra-orgs.values.buckets)
}
