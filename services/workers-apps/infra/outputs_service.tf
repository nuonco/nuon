output "orgs_ecr_access_role_arn" {
  value = module.orgs_ecr_access_role.iam_role_arn
}

output "orgs_k8s" {
  sensitive = true
  value = {
    ca_data         = data.tfe_outputs.infra-orgs.values.k8s.ca_data
    public_endpoint = data.tfe_outputs.infra-orgs.values.k8s.public_endpoint
    cluster_id      = data.tfe_outputs.infra-orgs.values.k8s.cluster_id
    role_arn        = data.tfe_outputs.infra-orgs.values.k8s.access_role_arns["eks-workers-apps"]
  }
}

output "orgs_role_templates" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.org_iam_role_name_templates)
}

output "waypoint_server" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.waypoint)
}

output "buckets" {
  value = nonsensitive(data.tfe_outputs.infra-orgs.values.buckets)
}
