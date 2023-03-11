output "orgs_ecr_access_role_arn" {
  value = module.orgs_ecr_access_role.iam_role_arn
}

output "orgs_role_templates" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.org_iam_role_name_templates)
}
