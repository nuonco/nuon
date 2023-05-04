output "buckets" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.buckets)
}

output "org_iam_role_name_templates" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.org_iam_role_name_templates)
}

output "orgs_k8s" {
  sensitive = true
  value = {
    ca_data         = data.tfe_outputs.infra-orgs.values.k8s.ca_data
    public_endpoint = data.tfe_outputs.infra-orgs.values.k8s.public_endpoint
    cluster_id      = data.tfe_outputs.infra-orgs.values.k8s.cluster_id
    role_arn        = data.tfe_outputs.infra-orgs.values.k8s.access_role_arns["eks-workers-installs"]
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
