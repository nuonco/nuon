output "buckets" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.buckets)
}

output "org_iam_role_name_templates" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.org_iam_role_name_templates)
}

output "orgs_k8s" {
  sensitive = false
  value = {
    ca_data         = nonsensitive(data.tfe_outputs.infra-orgs.values.k8s.ca_data)
    public_endpoint = nonsensitive(data.tfe_outputs.infra-orgs.values.k8s.public_endpoint)
    cluster_id      = nonsensitive(data.tfe_outputs.infra-orgs.values.k8s.cluster_id)
    role_arn        = nonsensitive(data.tfe_outputs.infra-orgs.values.k8s.access_role_arns["eks-workers-installs"])
  }
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

output "public_dns_access_role_arn" {
  value = module.public_dns_access_role.iam_role_arn
}

output "public_domain" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.public_domain)
}
