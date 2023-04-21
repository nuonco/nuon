output "buckets" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.buckets)
}

output "org_iam_role_name_templates" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.org_iam_role_name_templates)
}

# the following values are for interacting with the orgs k8s cluster, and while they currently are not being used should
# be migrated too when possible.
output "orgs_k8s" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.k8s)
}

output "sandbox" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.sandbox)
}

output "waypoint_server" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.waypoint)
}

output "orgs_iam_roles" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.iam_roles)
}
